package filtering

import expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

// WalkFunc is called for every expression while calling Walk.
// Return false to stop Walk.
type WalkFunc func(currExpr, parentExpr *expr.Expr) bool

// Walk an expression in depth-first order.
func Walk(fn WalkFunc, currExpr *expr.Expr) {
	walk(fn, currExpr, nil)
}

func walk(fn WalkFunc, currExpr, parentExpr *expr.Expr) {
	if fn == nil || currExpr == nil {
		return
	}
	if ok := fn(currExpr, parentExpr); !ok {
		return
	}
	switch v := currExpr.ExprKind.(type) {
	case *expr.Expr_ConstExpr, *expr.Expr_IdentExpr:
		// Nothing to do here.
	case *expr.Expr_SelectExpr:
		walk(fn, v.SelectExpr.Operand, currExpr)
	case *expr.Expr_CallExpr:
		walk(fn, v.CallExpr.Target, currExpr)
		for _, arg := range v.CallExpr.Args {
			walk(fn, arg, currExpr)
		}
	case *expr.Expr_ListExpr:
		for _, el := range v.ListExpr.Elements {
			walk(fn, el, currExpr)
		}
	case *expr.Expr_StructExpr:
		for _, entry := range v.StructExpr.Entries {
			walk(fn, entry.Value, currExpr)
		}
	case *expr.Expr_ComprehensionExpr:
		walk(fn, v.ComprehensionExpr.IterRange, currExpr)
		walk(fn, v.ComprehensionExpr.AccuInit, currExpr)
		walk(fn, v.ComprehensionExpr.LoopCondition, currExpr)
		walk(fn, v.ComprehensionExpr.LoopStep, currExpr)
		walk(fn, v.ComprehensionExpr.Result, currExpr)
	}
}
