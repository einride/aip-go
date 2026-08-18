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
	argument := val.Any()
	// Wildcards apply to declared string columns only. Every other type has a
	// parsed value that an asterisk cannot meaningfully match against.
	if text, ok := argument.(string); ok && c.schema.IsString(fld) {
		if (op == expr.Equal || op == expr.NotEqual) && hasWildcard(text) {
			return like(column, likePattern(text, false), op == expr.NotEqual), nil
		}
		argument = literalText(text)
	}
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
	// Membership is exact, so an asterisk here is data rather than a wildcard,
	// but the escape still resolves so `\*` reaches the column as `*`.
	text = literalText(text)
	if negated {
		return sq.Expr("NOT (? = ANY("+column+"))", text), nil
	}
	return sq.Expr("? = ANY("+column+")", text), nil
}

func (c *FilterCompiler) text(val expr.Value, negated bool) (sq.Sqlizer, error) {
	term, ok := val.Any().(string)
	if !ok {
		return nil, fmt.Errorf("invalid text search value")
	}
	// A term that names its own wildcards is honoured as written; one that does
	// not falls back to matching anywhere, which is what a bare search term
	// means.
	pattern := likePattern(term, !hasWildcard(term))

	fields := c.schema.SearchFields()
	predicates := make([]sq.Sqlizer, 0, len(fields))
	for _, field := range fields {
		predicates = append(predicates, like(c.fields[field], pattern, negated))
	}
	if negated {
		return sq.And(predicates), nil
	}
	return sq.Or(predicates), nil
}

// likeEscape is the escape character declared on every generated LIKE. User
// data never reaches the pattern unescaped, so a value containing %, _, or a
// backslash matches literally instead of being rejected or silently widened.
const likeEscape = '\\'

// hasWildcard reports whether val holds an unescaped asterisk.
//
// Bytes are enough: a wildcard, an escape, and the LIKE metacharacters are all
// ASCII, and no byte of a multi-byte rune can collide with them.
func hasWildcard(val string) bool {
	for i := 0; i < len(val); i++ {
		switch val[i] {
		case '\\':
			i++ // whatever follows is escaped, wildcard or not
		case '*':
			return true
		}
	}
	return false
}

// literalText resolves the pattern escape without giving the asterisk any
// meaning, so `\*` becomes `*`. It is what a comparison uses when the value
// holds no wildcard and therefore compiles to equality rather than LIKE.
func literalText(val string) string {
	if !strings.ContainsRune(val, likeEscape) {
		return val
	}
	var b strings.Builder
	b.Grow(len(val))
	for i := 0; i < len(val); i++ {
		if val[i] == '\\' && i+1 < len(val) {
			i++
		}
		b.WriteByte(val[i])
	}
	return b.String()
}

// likePattern renders val as a SQL LIKE pattern: an unescaped asterisk becomes
// %, `\*` becomes a literal asterisk, and every LIKE metacharacter carried by
// the data is escaped with likeEscape. When contains is set the pattern also
// matches anywhere in the column, which is what an unqualified search term
// means.
//
// A value can express a literal asterisk or a literal backslash, but not a
// literal backslash immediately followed by a wildcard: string unescaping runs
// first, so `\\*` has already collapsed to `\*` by the time it arrives here.
func likePattern(val string, contains bool) string {
	var b strings.Builder
	b.Grow(len(val) + 2)
	if contains {
		b.WriteByte('%')
	}
	for i := 0; i < len(val); i++ {
		switch c := val[i]; {
		case c == '\\' && i+1 < len(val):
			i++
			writeLikeLiteral(&b, val[i])
		case c == '*':
			b.WriteByte('%')
		default:
			writeLikeLiteral(&b, c)
		}
	}
	if contains {
		b.WriteByte('%')
	}
	return b.String()
}

func writeLikeLiteral(b *strings.Builder, c byte) {
	if c == '%' || c == '_' || c == likeEscape {
		b.WriteByte(likeEscape)
	}
	b.WriteByte(c)
}

// like builds a LIKE with an explicit escape declaration. Squirrel's sq.Like
// cannot carry one, and the column is already checked against the identifier
// rule by NewFilterCompiler, the same trust boundary ordering relies on.
func like(column, pattern string, negated bool) sq.Sqlizer {
	if negated {
		return sq.Expr(column+" NOT LIKE ? ESCAPE '\\'", pattern)
	}
	return sq.Expr(column+" LIKE ? ESCAPE '\\'", pattern)
}
