package filter

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"time"

	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/internal/ident"
)

// Limits bounds the complexity of filters accepted by an endpoint.
type Limits struct {
	MaxBytes int
	MaxDepth int
	MaxNodes int
}

type typeKind uint8

const (
	boolType typeKind = iota + 1
	intType
	uintType
	floatType
	stringType
	timeType
	durationType
	enumType
	stringListType
)

// Type describes the values accepted by a filter field.
type Type struct {
	kind   typeKind
	values []string
}

var (
	TypeBool     = Type{kind: boolType}
	TypeInt      = Type{kind: intType}
	TypeUint     = Type{kind: uintType}
	TypeFloat    = Type{kind: floatType}
	TypeString   = Type{kind: stringType}
	TypeTime     = Type{kind: timeType}
	TypeDuration = Type{kind: durationType}

	// TypeStringList declares a repeated string column. It accepts the has
	// operator and nothing else.
	TypeStringList = Type{kind: stringListType}
)

// Enum creates a case-sensitive enum filter type.
func Enum(vals ...string) Type {
	return Type{kind: enumType, values: slices.Clone(vals)}
}

// Field declares one API field accepted by an endpoint filter.
type Field struct {
	Name   string
	Type   Type
	Search bool
}

// NewField declares one API field accepted by an endpoint filter. A searchable
// field must be a string; NewSchema rejects any other type.
func NewField(name string, typ Type, search bool) Field {
	return Field{Name: name, Type: typ, Search: search}
}

// Schema validates parsed filters for one endpoint.
type Schema struct {
	fields      map[string]Field
	fieldOrder  []string
	searchOrder []string
	limits      Limits
	disabled    bool
}

// Disabled declares that an endpoint accepts only an empty filter.
func Disabled() *Schema {
	return &Schema{disabled: true}
}

// NewSchema validates and copies a filter declaration.
func NewSchema(fields []Field, limits Limits) (*Schema, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("filter: schema has no fields")
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = maxBytes
	}
	if limits.MaxDepth == 0 {
		limits.MaxDepth = maxDepth
	}
	if limits.MaxNodes == 0 {
		limits.MaxNodes = maxNodes
	}
	if limits.MaxBytes < 0 || limits.MaxDepth < 0 || limits.MaxNodes < 0 {
		return nil, fmt.Errorf("filter: limits must be positive")
	}

	schema := &Schema{
		fields:      make(map[string]Field, len(fields)),
		fieldOrder:  make([]string, 0, len(fields)),
		searchOrder: make([]string, 0, len(fields)),
		limits:      limits,
	}

	for _, fld := range fields {
		if !ident.ValidPath(fld.Name) {
			return nil, fmt.Errorf("filter: invalid field %q", fld.Name)
		}
		if _, exists := schema.fields[fld.Name]; exists {
			return nil, fmt.Errorf("filter: duplicate field %q", fld.Name)
		}
		if err := validateType(fld.Type); err != nil {
			return nil, fmt.Errorf("filter: field %q: %w", fld.Name, err)
		}
		if fld.Search && fld.Type.kind != stringType {
			return nil, fmt.Errorf("filter: searchable field %q must be a string", fld.Name)
		}
		fld.Type.values = slices.Clone(fld.Type.values)
		schema.fields[fld.Name] = fld
		schema.fieldOrder = append(schema.fieldOrder, fld.Name)
		if fld.Search {
			schema.searchOrder = append(schema.searchOrder, fld.Name)
		}
	}
	return schema, nil
}

func validateType(typ Type) error {
	if typ.kind < boolType || typ.kind > stringListType {
		return fmt.Errorf("invalid filter type")
	}
	if typ.kind != enumType {
		return nil
	}
	if len(typ.values) == 0 {
		return fmt.Errorf("enum has no values")
	}
	seen := make(map[string]struct{}, len(typ.values))
	for _, val := range typ.values {
		if val == "" {
			return fmt.Errorf("enum has an empty value")
		}
		if _, exists := seen[val]; exists {
			return fmt.Errorf("enum has duplicate value %q", val)
		}
		seen[val] = struct{}{}
	}
	return nil
}

// Initialized reports whether s was created by NewSchema or Disabled.
// NewSchema rejects an empty field list, so a schema that declares nothing and
// is not disabled cannot have come from a constructor.
func (s *Schema) Initialized() bool {
	return s != nil && (s.disabled || len(s.fields) > 0)
}

// Fields returns the declared fields in declaration order.
func (s *Schema) Fields() []Field {
	fields := make([]Field, 0, len(s.fieldOrder))
	for _, name := range s.fieldOrder {
		fld := s.fields[name]
		fld.Type.values = slices.Clone(fld.Type.values)
		fields = append(fields, fld)
	}
	return fields
}

// SearchFields returns searchable field names in declaration order.
func (s *Schema) SearchFields() []string {
	return slices.Clone(s.searchOrder)
}

// IsString reports whether name is a declared string field.
func (s *Schema) IsString(name string) bool {
	field, ok := s.fields[name]
	return ok && field.Type.kind == stringType
}

// Validate checks a parsed filter against the schema and returns a typed copy.
func (s *Schema) Validate(f Filter) (Filter, error) {
	if !s.Initialized() {
		return Filter{}, fmt.Errorf("filter: uninitialized schema")
	}
	if f.Expr == nil {
		return Filter{}, nil
	}
	if s.disabled {
		return Filter{}, fmt.Errorf("filter: filtering is disabled")
	}
	nodes := 0
	e, err := s.validateExpr(f.Expr, 1, &nodes)
	if err != nil {
		return Filter{}, err
	}
	validated := Filter{Expr: e}
	if len(validated.String()) > s.limits.MaxBytes {
		return Filter{}, fmt.Errorf("filter: filter is too large")
	}
	return validated, nil
}

func (s *Schema) validateExpr(e expr.Expr, depth int, nodes *int) (expr.Expr, error) {
	*nodes++
	if depth > s.limits.MaxDepth {
		return nil, fmt.Errorf("filter: nesting is too deep")
	}
	if *nodes > s.limits.MaxNodes {
		return nil, fmt.Errorf("filter: too many expressions")
	}
	switch node := e.(type) {
	case expr.LogicalExpr:
		left, err := s.validateExpr(node.Left, depth+1, nodes)
		if err != nil {
			return nil, err
		}
		right, err := s.validateExpr(node.Right, depth+1, nodes)
		if err != nil {
			return nil, err
		}
		if node.Op != expr.And && node.Op != expr.Or && node.Op != expr.Seq {
			return nil, fmt.Errorf("filter: invalid logical operator")
		}
		return expr.LogicalExpr{Left: left, Op: node.Op, Right: right}, nil
	case expr.NotExpr:
		inner, err := s.validateExpr(node.Expr, depth+1, nodes)
		if err != nil {
			return nil, err
		}
		return expr.NotExpr{Expr: inner}, nil
	case expr.ComparisonExpr:
		field, ok := s.fields[node.Field]
		if !ok {
			return nil, fmt.Errorf("filter: unknown field %q", node.Field)
		}
		value, err := coerceValue(node.Value, field.Type)
		if err != nil {
			return nil, fmt.Errorf("filter: field %q: %w", node.Field, err)
		}
		if !opAllowed(node.Op, field.Type, value) {
			return nil, fmt.Errorf(
				"filter: operator %q is not allowed for field %q",
				node.Op.String(),
				node.Field,
			)
		}
		return expr.ComparisonExpr{Field: node.Field, Op: node.Op, Value: value}, nil
	case expr.TextExpr:
		if len(s.searchOrder) == 0 {
			return nil, fmt.Errorf("filter: unqualified text search is disabled")
		}
		if !node.V.Valid() || node.V.Kind() == expr.NullKind {
			return nil, fmt.Errorf("filter: invalid text search value")
		}
		if node.V.Kind() == expr.StringKind {
			return expr.TextExpr{V: node.V}, nil
		}
		return expr.TextExpr{V: expr.StringValue(node.V.Canonical())}, nil
	default:
		return nil, fmt.Errorf("filter: invalid expression")
	}
}

func opAllowed(op expr.ComparisonOp, typ Type, val expr.Value) bool {
	if op < expr.Equal || op > expr.Has {
		return false
	}
	// A repeated column supports membership only, and every other column
	// supports everything but membership.
	if typ.kind == stringListType {
		return op == expr.Has && val.Kind() == expr.StringKind
	}
	if op == expr.Has {
		return false
	}
	if val.Kind() == expr.NullKind || typ.kind == boolType || typ.kind == enumType {
		return op == expr.Equal || op == expr.NotEqual
	}
	return true
}

func coerceValue(val expr.Value, typ Type) (expr.Value, error) {
	if !val.Valid() {
		return expr.Value{}, fmt.Errorf("invalid value")
	}
	if val.Kind() == expr.NullKind {
		return val, nil
	}
	switch typ.kind {
	case boolType:
		if val.Kind() == expr.BoolKind {
			return val, nil
		}
	case intType:
		switch v := val.Any().(type) {
		case int64:
			return val, nil
		case uint64:
			if v <= math.MaxInt64 {
				return expr.Int64Value(int64(v)), nil
			}
		case string:
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				return expr.Int64Value(parsed), nil
			}
		}
	case uintType:
		switch v := val.Any().(type) {
		case uint64:
			return val, nil
		case int64:
			if v >= 0 {
				return expr.Uint64Value(uint64(v)), nil
			}
		case string:
			if parsed, err := strconv.ParseUint(v, 10, 64); err == nil {
				return expr.Uint64Value(parsed), nil
			}
		}
	case floatType:
		switch v := val.Any().(type) {
		case float64:
			return val, nil
		case int64:
			return expr.Float64Value(float64(v)), nil
		case uint64:
			return expr.Float64Value(float64(v)), nil
		case string:
			parsed, err := strconv.ParseFloat(v, 64)
			if err == nil && !math.IsInf(parsed, 0) && !math.IsNaN(parsed) {
				return expr.Float64Value(parsed), nil
			}
		}
	case stringType, stringListType:
		if val.Kind() == expr.StringKind {
			return val, nil
		}
	case timeType:
		if val.Kind() == expr.TimeKind {
			return val, nil
		}
		if v, ok := val.Any().(string); ok {
			if parsed, err := time.Parse(time.RFC3339Nano, v); err == nil {
				return expr.TimeValue(parsed), nil
			}
		}
	case durationType:
		if val.Kind() == expr.DurationKind {
			return val, nil
		}
		if v, ok := val.Any().(string); ok {
			if parsed, err := time.ParseDuration(v); err == nil {
				return expr.DurationValue(parsed), nil
			}
		}
	case enumType:
		if v, ok := val.Any().(string); ok && slices.Contains(typ.values, v) {
			return val, nil
		}
	}
	return expr.Value{}, fmt.Errorf("value does not match the declared type")
}
