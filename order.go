package flop

import (
	"slices"
	"strings"

	"github.com/ysomad/flop/internal/ident"
)

// OrderField is one field in an AIP-132 order_by expression.
type OrderField struct {
	Path string
	Desc bool
}

// Asc creates an ascending order field.
func Asc(path string) OrderField { return OrderField{Path: path} }

// Desc creates a descending order field.
func Desc(path string) OrderField { return OrderField{Path: path, Desc: true} }

// SQL returns the field as an explicit SQL ORDER BY expression.
// Path is safe only after it has been checked against a trusted declaration.
func (f OrderField) SQL() string {
	if f.Desc {
		return f.Path + " DESC"
	}
	return f.Path + " ASC"
}

// OrderBy is an ordered list of fields.
//
// The zero value represents no client-provided ordering.
type OrderBy struct {
	Fields []OrderField
}

func (o OrderBy) clone() OrderBy {
	return OrderBy{Fields: slices.Clone(o.Fields)}
}

// ParseOrderBy parses and validates an AIP-132 order_by string.
func ParseOrderBy(input string) (OrderBy, error) {
	if strings.TrimSpace(input) == "" {
		return OrderBy{}, nil
	}

	parts := strings.Split(input, ",")
	fields := make([]OrderField, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		words := strings.Fields(part)
		if len(words) == 0 || len(words) > 2 {
			return OrderBy{}, errorf(ErrInvalidOrder, "malformed order_by field")
		}
		if !ident.ValidPath(words[0]) {
			return OrderBy{}, errorf(ErrInvalidOrder, "malformed order_by path")
		}
		if _, ok := seen[words[0]]; ok {
			return OrderBy{}, errorf(ErrInvalidOrder, "duplicate order_by path")
		}

		field := OrderField{Path: words[0]}
		if len(words) == 2 {
			switch words[1] {
			case "asc":
			case "desc":
				field.Desc = true
			default:
				return OrderBy{}, errorf(ErrInvalidOrder, "direction must be asc or desc")
			}
		}
		seen[field.Path] = struct{}{}
		fields = append(fields, field)
	}

	return OrderBy{Fields: fields}, nil
}

// String returns the canonical AIP-132 representation of o.
func (o OrderBy) String() string {
	if len(o.Fields) == 0 {
		return ""
	}
	var b strings.Builder
	for i, field := range o.Fields {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(field.Path)
		if field.Desc {
			b.WriteString(" desc")
		}
	}
	return b.String()
}
