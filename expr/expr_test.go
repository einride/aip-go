package expr_test

import (
	"testing"
	"time"

	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/internal/assert"
)

func TestCanonical(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		e    expr.Expr
		want string
	}{
		{name: "nil", want: ""},
		{
			name: "logical expression",
			e: expr.LogicalExpr{
				Left: expr.ComparisonExpr{
					Field: "name",
					Op:    expr.Equal,
					Value: expr.StringValue("book"),
				},
				Op: expr.And,
				Right: expr.NotExpr{Expr: expr.ComparisonExpr{
					Field: "deleted",
					Op:    expr.Equal,
					Value: expr.BoolValue(true),
				}},
			},
			want: `(name = "book" AND NOT (deleted = true))`,
		},
		{
			name: "sequence expression",
			e: expr.LogicalExpr{
				Left: expr.TextExpr{V: expr.StringValue("blue")}, Op: expr.Seq,
				Right: expr.TextExpr{V: expr.StringValue("chair")},
			},
			want: `("blue" AND "chair")`,
		},
		{
			name: "unknown logical operator defaults to and",
			e: expr.LogicalExpr{
				Left:  expr.TextExpr{V: expr.StringValue("blue")},
				Right: expr.TextExpr{V: expr.StringValue("chair")},
			},
			want: `("blue" AND "chair")`,
		},
		{
			name: "or expression",
			e: expr.LogicalExpr{
				Left:  expr.TextExpr{V: expr.StringValue("blue")},
				Op:    expr.Or,
				Right: expr.TextExpr{V: expr.StringValue("chair")},
			},
			want: `("blue" OR "chair")`,
		},
		{
			name: "has",
			e:    expr.ComparisonExpr{Field: "tags", Op: expr.Has, Value: expr.StringValue("blue")},
			want: `tags : "blue"`,
		},
		{name: "null value", e: expr.TextExpr{V: expr.NullValue()}, want: "null"},
		{name: "int64 value", e: expr.TextExpr{V: expr.Int64Value(-1)}, want: "-1"},
		{name: "uint64 value", e: expr.TextExpr{V: expr.Uint64Value(1)}, want: "1"},
		{name: "float64 value", e: expr.TextExpr{V: expr.Float64Value(1.5)}, want: "1.5"},
		{name: "bytes value", e: expr.TextExpr{V: expr.BytesValue([]byte("value\n"))}, want: `"value\n"`},
		{name: "time value", e: expr.TextExpr{V: expr.TimeValue(time.Unix(1, 0))}, want: "1970-01-01T00:00:01Z"},
		{name: "duration value", e: expr.TextExpr{V: expr.DurationValue(1500 * time.Millisecond)}, want: "1.5s"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, expr.Canonical(test.e))
		})
	}
}

func TestComparisonOp_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		operator expr.ComparisonOp
		want     string
	}{
		{name: "equal", operator: expr.Equal, want: "="},
		{name: "not equal", operator: expr.NotEqual, want: "!="},
		{name: "less", operator: expr.Less, want: "<"},
		{name: "less or equal", operator: expr.LessOrEqual, want: "<="},
		{name: "greater", operator: expr.Greater, want: ">"},
		{name: "greater or equal", operator: expr.GreaterOrEqual, want: ">="},
		{name: "has", operator: expr.Has, want: ":"},
		{name: "unknown", operator: expr.ComparisonOp(255), want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.operator.String())
		})
	}
}

func TestComparisonOp_Invert(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		op   expr.ComparisonOp
		want expr.ComparisonOp
	}{
		{name: "equal", op: expr.Equal, want: expr.NotEqual},
		{name: "not equal", op: expr.NotEqual, want: expr.Equal},
		{name: "less", op: expr.Less, want: expr.GreaterOrEqual},
		{name: "less or equal", op: expr.LessOrEqual, want: expr.Greater},
		{name: "greater", op: expr.Greater, want: expr.LessOrEqual},
		{name: "greater or equal", op: expr.GreaterOrEqual, want: expr.Less},
		{name: "has has no inverse", op: expr.Has},
		{name: "unknown", op: expr.ComparisonOp(255)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.op.Invert())
		})
	}
}
