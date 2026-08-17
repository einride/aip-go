package sqflop

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/ysomad/flop"
	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/internal/ident"
)

// FilterCompiler compiles validated filters into Squirrel predicates.
type FilterCompiler struct {
	schema *flop.FilterSchema
	fields map[string]string
}

// NewFilterCompiler maps API filter fields onto SQL columns. Fields absent from
// fields use their own name as the column, so only columns that differ — a
// table-qualified name, or a rename — need an entry. A nil map leaves every
// field on its own name.
func NewFilterCompiler(schema *flop.FilterSchema, fields map[string]string) (*FilterCompiler, error) {
	if _, err := schema.Validate(flop.Filter{}); err != nil {
		return nil, fmt.Errorf("sqflop: %w", err)
	}

	comp := &FilterCompiler{schema: schema, fields: make(map[string]string, len(fields))}
	declared := schema.Fields()
	known := make(map[string]struct{}, len(declared))

	for _, fld := range declared {
		known[fld.Name] = struct{}{}

		col, ok := fields[fld.Name]
		if !ok {
			col = fld.Name
		}

		if !ident.ValidPath(col) {
			return nil, fmt.Errorf("sqflop: invalid SQL column %q", col)
		}

		comp.fields[fld.Name] = col
	}

	for name := range fields {
		if _, ok := known[name]; !ok {
			return nil, fmt.Errorf("sqflop: SQL column maps unknown filter field %q", name)
		}
	}

	return comp, nil
}

// Where adds filter to base. Compilation errors surface when ToSql is called.
func (c *FilterCompiler) Where(base sq.SelectBuilder, f flop.Filter) sq.SelectBuilder {
	validated, err := c.schema.Validate(f)
	if err != nil {
		return wrapErr(base, "filter: %w", err)
	}
	if validated.Expr == nil {
		return base
	}

	predicate, err := c.compile(validated.Expr, false)
	if err != nil {
		return wrapErr(base, "filter: %w", err)
	}

	return base.Where(predicate)
}

func (c *FilterCompiler) compile(e expr.Expr, negated bool) (sq.Sqlizer, error) {
	switch node := e.(type) {
	case expr.LogicalExpr:
		left, err := c.compile(node.Left, negated)
		if err != nil {
			return nil, err
		}
		right, err := c.compile(node.Right, negated)
		if err != nil {
			return nil, err
		}
		isOr := node.Op == expr.Or
		if negated {
			isOr = !isOr
		}
		if isOr {
			return sq.Or{left, right}, nil
		}
		return sq.And{left, right}, nil
	case expr.NotExpr:
		return c.compile(node.Expr, !negated)
	case expr.ComparisonExpr:
		// Has has no inverted comparator, so it is negated structurally rather
		// than by flipping the operator.
		if node.Op == expr.Has {
			return c.has(node.Field, node.Value, negated)
		}
		operator := node.Op
		if negated {
			operator = operator.Invert()
		}
		return c.comparison(node.Field, operator, node.Value)
	case expr.TextExpr:
		return c.text(node.V, negated)
	default:
		return nil, fmt.Errorf("invalid expression")
	}
}

func (c *FilterCompiler) comparison(fld string, op expr.ComparisonOp, val expr.Value) (sq.Sqlizer, error) {
	column, ok := c.fields[fld]
	if !ok {
		return nil, fmt.Errorf("unknown filter field %q", fld)
	}
	if c.schema.IsString(fld) && (op == expr.Equal || op == expr.NotEqual) {
		text, _ := val.Any().(string)
		if val.Kind() == expr.StringKind && hasBoundaryWildcard(text) {
			pattern, err := filterPattern(text, false)
			if err != nil {
				return nil, err
			}
			if op == expr.NotEqual {
				return sq.NotLike{column: pattern}, nil
			}
			return sq.Like{column: pattern}, nil
		}
	}
	argument := val.Any()
	switch op {
	case expr.Equal:
		return sq.Eq{column: argument}, nil
	case expr.NotEqual:
		return sq.NotEq{column: argument}, nil
	case expr.Less:
		return sq.Lt{column: argument}, nil
	case expr.LessOrEqual:
		return sq.LtOrEq{column: argument}, nil
	case expr.Greater:
		return sq.Gt{column: argument}, nil
	case expr.GreaterOrEqual:
		return sq.GtOrEq{column: argument}, nil
	default:
		return nil, fmt.Errorf("invalid comparison operator")
	}
}

// has compiles the AIP-160 has operator to Postgres array membership.
//
// The column is already checked against the identifier rule by
// NewFilterCompiler, which is the same trust boundary that lets ordering
// interpolate column names; the value stays a bound argument.
func (c *FilterCompiler) has(fld string, val expr.Value, negated bool) (sq.Sqlizer, error) {
	column, ok := c.fields[fld]
	if !ok {
		return nil, fmt.Errorf("unknown filter field %q", fld)
	}
	text, ok := val.Any().(string)
	if !ok {
		return nil, fmt.Errorf("invalid has value")
	}
	if negated {
		return sq.Expr("NOT (? = ANY("+column+"))", text), nil
	}
	return sq.Expr("? = ANY("+column+")", text), nil
}

func (c *FilterCompiler) text(val expr.Value, negated bool) (sq.Sqlizer, error) {
	text, ok := val.Any().(string)
	if !ok {
		return nil, fmt.Errorf("invalid text search value")
	}
	pattern, err := filterPattern(text, true)
	if err != nil {
		return nil, err
	}
	fields := c.schema.SearchFields()
	predicates := make([]sq.Sqlizer, 0, len(fields))
	for _, field := range fields {
		column := c.fields[field]
		if negated {
			predicates = append(predicates, sq.NotLike{column: pattern})
		} else {
			predicates = append(predicates, sq.Like{column: pattern})
		}
	}
	if negated {
		return sq.And(predicates), nil
	}
	return sq.Or(predicates), nil
}

func filterPattern(val string, contains bool) (string, error) {
	leading := strings.HasPrefix(val, "*")
	trailing := strings.HasSuffix(val, "*")
	if leading {
		val = strings.TrimPrefix(val, "*")
	}
	if trailing && val != "" {
		val = strings.TrimSuffix(val, "*")
	}
	if strings.ContainsAny(val, "%_\\") {
		return "", ErrUnsupportedPattern
	}
	if contains || leading {
		val = "%" + val
	}
	if contains || trailing {
		val += "%"
	}
	return val, nil
}

func hasBoundaryWildcard(v string) bool {
	return strings.HasPrefix(v, "*") || strings.HasSuffix(v, "*")
}
