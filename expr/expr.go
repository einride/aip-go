package expr

import "fmt"

// Canonical renders e as AIP-160 source. A nil expression renders as the empty
// string.
//
// This is the only way to reach a node's rendering from outside the package,
// because canonical is unexported. The output is what callers fingerprint, so
// it is a compatibility surface rather than a debug format.
func Canonical(e Expr) string {
	if e == nil {
		return ""
	}
	return e.canonical()
}

// Expr is a node in a filter expression tree.
//
// The set of expressions is closed: canonical is unexported, so only this
// package can implement Expr. Consumers rely on that to switch over the node
// types exhaustively.
type Expr interface {
	canonical() string
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

func (e LogicalExpr) canonical() string {
	// Seq renders as AND: whitespace sequencing is a conjunction, and the
	// canonical form deliberately does not distinguish the two spellings.
	op := " AND "
	if e.Op == Or {
		op = " OR "
	}
	return fmt.Sprintf("(%s%s%s)", e.Left.canonical(), op, e.Right.canonical())
}

// NotExpr negates Expr.
type NotExpr struct {
	Expr Expr
}

func (e NotExpr) canonical() string {
	return fmt.Sprintf("NOT (%s)", e.Expr.canonical())
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

func (e ComparisonExpr) canonical() string {
	return fmt.Sprintf("%s %s %s", e.Field, e.Op.String(), e.Value.Canonical())
}

// TextExpr is an unqualified text-search expression.
type TextExpr struct {
	V Value
}

func (e TextExpr) canonical() string {
	return e.V.Canonical()
}
