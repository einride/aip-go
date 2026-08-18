package expr_test

import (
	"strconv"
	"testing"

	"github.com/ysomad/flop/expr"
)

// chain builds a left-leaning conjunction of n comparisons, the shape a flat
// "a = 0 AND a = 1 AND ..." filter parses into.
func chain(n int) expr.Expr {
	var e expr.Expr = expr.ComparisonExpr{Field: "a", Op: expr.Equal, Value: expr.Int64Value(0)}
	for i := 1; i < n; i++ {
		e = expr.LogicalExpr{
			Left:  e,
			Op:    expr.And,
			Right: expr.ComparisonExpr{Field: "a", Op: expr.Equal, Value: expr.Int64Value(int64(i))},
		}
	}
	return e
}

// BenchmarkCanonical guards against a return to per-node string building, which
// made rendering quadratic in the node count.
func BenchmarkCanonical(b *testing.B) {
	for _, n := range []int{16, 64, 256} {
		e := chain(n)
		b.Run("nodes="+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = expr.Canonical(e)
			}
		})
	}
}

func BenchmarkValueCanonical(b *testing.B) {
	b.ReportAllocs()
	v := expr.StringValue("pay_10")
	for range b.N {
		_ = v.Canonical()
	}
}
