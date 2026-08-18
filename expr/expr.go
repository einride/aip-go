package expr

import "strings"

// Canonical renders e as AIP-160 source. A nil expression renders as the empty
// string.
//
// This is the only way to reach a node's rendering from outside the package,
// because appendCanonical is unexported. The output is what callers
// fingerprint, so it is a compatibility surface rather than a debug format.
func Canonical(e Expr) string {
	if e == nil {
		return ""
	}
	var b strings.Builder
	e.appendCanonical(&b)
	return b.String()
}

// Expr is a node in a filter expression tree.
//
// The set of expressions is closed: appendCanonical is unexported, so only this
// package can implement Expr. Consumers rely on that to switch over the node
// types exhaustively.
//
// Nodes append into a caller-owned builder rather than returning a string, so
// rendering a tree of n nodes stays linear. Returning a string per node made
// every level recopy the whole accumulated rendering.
type Expr interface {
	appendCanonical(*strings.Builder)
}

// LogicalOp joins two filter expressions.
type LogicalOp uint8

const (
	And LogicalOp = iota + 1
	Or
	Seq
)

// LogicalExpr joins Left and Right with Operator.
//
// Sequence represents AIP-160 whitespace sequencing and is intentionally
// distinct from an explicit AND.
type LogicalExpr struct {
	Left  Expr
	Op    LogicalOp
	Right Expr
}

func (e LogicalExpr) appendCanonical(b *strings.Builder) {
	// Seq renders as AND: whitespace sequencing is a conjunction, and the
	// canonical form deliberately does not distinguish the two spellings.
	op := " AND "
	if e.Op == Or {
		op = " OR "
	}
	b.WriteByte('(')
	e.Left.appendCanonical(b)
	b.WriteString(op)
	e.Right.appendCanonical(b)
	b.WriteByte(')')
}

// NotExpr negates Expr.
type NotExpr struct {
	Expr Expr
}

func (e NotExpr) appendCanonical(b *strings.Builder) {
	b.WriteString("NOT (")
	e.Expr.appendCanonical(b)
	b.WriteByte(')')
}

// ComparisonOp compares a field with a literal value.
type ComparisonOp uint8

const (
	Equal ComparisonOp = iota + 1
	NotEqual
	Less
	LessOrEqual
	Greater
	GreaterOrEqual
	Has
)

func (o ComparisonOp) String() string {
	switch o {
	case Equal:
		return "="
	case NotEqual:
		return "!="
	case Less:
		return "<"
	case LessOrEqual:
		return "<="
	case Greater:
		return ">"
	case GreaterOrEqual:
		return ">="
	case Has:
		return ":"
	default:
		return ""
	}
}

// Invert returns the comparison that is true when o is false.
//
// Has has no inverted comparator and returns the zero value; negating it is
// the compiler's job.
func (o ComparisonOp) Invert() ComparisonOp {
	switch o {
	case Equal:
		return NotEqual
	case NotEqual:
		return Equal
	case Less:
		return GreaterOrEqual
	case LessOrEqual:
		return Greater
	case Greater:
		return LessOrEqual
	case GreaterOrEqual:
		return Less
	default:
		return 0
	}
}

// ComparisonExpr compares Field with Value.
type ComparisonExpr struct {
	Field string
	Op    ComparisonOp
	Value Value
}

func (e ComparisonExpr) appendCanonical(b *strings.Builder) {
	b.WriteString(e.Field)
	b.WriteByte(' ')
	b.WriteString(e.Op.String())
	b.WriteByte(' ')
	e.Value.appendCanonical(b)
}

// TextExpr is an unqualified text-search expression.
type TextExpr struct {
	V Value
}

func (e TextExpr) appendCanonical(b *strings.Builder) {
	e.V.appendCanonical(b)
}
