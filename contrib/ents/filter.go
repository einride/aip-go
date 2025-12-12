package ents

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"go.einride.tech/aip/filtering"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// ApplyFilter builds a SQL selector from a filtering.Filter.
// Example: name="value" AND age>18
// More detail in [AIP-160](https://google.aip.dev/160).
func ApplyFilter(filter filtering.Filter) func(*sql.Selector) {
	return func(s *sql.Selector) {
		if filter.CheckedExpr == nil || filter.CheckedExpr.Expr == nil {
			return
		}
		predicate, err := exprToPredicate(s, filter.CheckedExpr.Expr)
		if err != nil || predicate == nil {
			// Fall back to a no-op if the filter expression cannot be translated.
			return
		}
		s.Where(predicate)
	}
}

// constToValue converts a Constant expression to its Go value.
func constToValue(constExpr *expr.Constant) (any, error) {
	switch constExpr.ConstantKind.(type) {
	case *expr.Constant_BoolValue:
		return constExpr.GetBoolValue(), nil
	case *expr.Constant_StringValue:
		return constExpr.GetStringValue(), nil
	case *expr.Constant_Int64Value:
		return constExpr.GetInt64Value(), nil
	case *expr.Constant_DoubleValue:
		return constExpr.GetDoubleValue(), nil
	case *expr.Constant_DurationValue:
		if d := constExpr.GetDurationValue(); d != nil {
			return d.AsDuration(), nil
		}
	case *expr.Constant_TimestampValue:
		if ts := constExpr.GetTimestampValue(); ts != nil {
			return ts.AsTime(), nil
		}
	}
	return nil, fmt.Errorf("unsupported constant expression %v", constExpr.ConstantKind)
}

// exprToPredicate converts an expression to a SQL predicate.
func exprToPredicate(sel *sql.Selector, e *expr.Expr) (*sql.Predicate, error) {
	if e == nil {
		return nil, fmt.Errorf("expression is nil")
	}

	switch kind := e.GetExprKind().(type) {
	case *expr.Expr_CallExpr:
		call := kind.CallExpr
		function := call.GetFunction()
		args := call.GetArgs()

		switch function {
		case filtering.FunctionAnd, filtering.FunctionFuzzyAnd:
			return combineLogical(sel, args, sql.And)
		case filtering.FunctionOr:
			return combineLogical(sel, args, sql.Or)
		case filtering.FunctionNot:
			if len(args) != 1 {
				return nil, fmt.Errorf("NOT expects 1 argument, got %d", len(args))
			}
			pred, err := exprToPredicate(sel, args[0])
			if err != nil {
				return nil, err
			}
			return sql.Not(pred), nil
		case filtering.FunctionEquals,
			filtering.FunctionNotEquals,
			filtering.FunctionLessThan,
			filtering.FunctionLessEquals,
			filtering.FunctionGreaterThan,
			filtering.FunctionGreaterEquals:
			return comparisonPredicate(sel, function, args)
		case filtering.FunctionHas:
			return hasPredicate(sel, args)
		default:
			return nil, fmt.Errorf("unsupported call expression: %s", function)
		}

	case *expr.Expr_ConstExpr:
		value, err := constToValue(kind.ConstExpr)
		if err != nil {
			return nil, err
		}
		boolean, ok := value.(bool)
		if !ok {
			return nil, fmt.Errorf("non-boolean constant cannot form predicate: %v", value)
		}
		if boolean {
			return sql.ExprP("TRUE"), nil
		}
		return sql.ExprP("FALSE"), nil

	case *expr.Expr_IdentExpr:
		// Interpret bare identifiers as checking for truthy/true columns.
		return sql.IsTrue(sel.C(kind.IdentExpr.GetName())), nil

	case *expr.Expr_SelectExpr:
		column, ok := exprToColumn(kind.SelectExpr)
		if !ok {
			return nil, fmt.Errorf("invalid select expression")
		}
		return sql.IsTrue(sel.C(column)), nil
	}

	return nil, fmt.Errorf("unsupported expression kind %T", e.GetExprKind())
}

// combineLogical combines multiple expressions into a single predicate using the provided combine function.
func combineLogical(sel *sql.Selector, args []*expr.Expr, combine func(...*sql.Predicate) *sql.Predicate) (*sql.Predicate, error) {
	if len(args) == 0 {
		return sql.ExprP("TRUE"), nil
	}
	preds := make([]*sql.Predicate, 0, len(args))
	for _, arg := range args {
		pred, err := exprToPredicate(sel, arg)
		if err != nil {
			return nil, err
		}
		preds = append(preds, pred)
	}
	return combine(preds...), nil
}

// comparisonPredicate builds a comparison predicate from two expressions.
func comparisonPredicate(sel *sql.Selector, function string, args []*expr.Expr) (*sql.Predicate, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%s expects 2 arguments, got %d", function, len(args))
	}

	leftCol, leftIsCol := exprToColumnExpr(args[0])
	rightCol, rightIsCol := exprToColumnExpr(args[1])

	switch {
	case leftIsCol && rightIsCol:
		return columnToColumnPredicate(sel, function, leftCol, rightCol)
	case leftIsCol:
		val, err := exprToValue(args[1])
		if err != nil {
			return nil, err
		}
		return columnToValuePredicate(sel, function, leftCol, val)
	case rightIsCol:
		val, err := exprToValue(args[0])
		if err != nil {
			return nil, err
		}
		return valueToColumnPredicate(sel, function, val, rightCol)
	default:
		return nil, fmt.Errorf("comparison requires at least one column operand")
	}
}

// hasPredicate builds a "has" predicate from two expressions.
func hasPredicate(sel *sql.Selector, args []*expr.Expr) (*sql.Predicate, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%s expects 2 arguments, got %d", filtering.FunctionHas, len(args))
	}
	column, ok := exprToColumnExpr(args[0])
	if !ok {
		return nil, fmt.Errorf("left operand of ':' must be a column")
	}
	value, err := exprToValue(args[1])
	if err != nil {
		return nil, err
	}
	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("':' right operand must resolve to string, got %T", value)
	}
	return sql.Contains(sel.C(column), str), nil
}

// columnToColumnPredicate builds a predicate comparing two columns.
func columnToColumnPredicate(sel *sql.Selector, function, left, right string) (*sql.Predicate, error) {
	switch function {
	case filtering.FunctionEquals:
		return sql.ColumnsEQ(sel.C(left), sel.C(right)), nil
	case filtering.FunctionNotEquals:
		return sql.ColumnsNEQ(sel.C(left), sel.C(right)), nil
	case filtering.FunctionLessThan:
		return sql.ColumnsLT(sel.C(left), sel.C(right)), nil
	case filtering.FunctionLessEquals:
		return sql.ColumnsLTE(sel.C(left), sel.C(right)), nil
	case filtering.FunctionGreaterThan:
		return sql.ColumnsGT(sel.C(left), sel.C(right)), nil
	case filtering.FunctionGreaterEquals:
		return sql.ColumnsGTE(sel.C(left), sel.C(right)), nil
	default:
		return nil, fmt.Errorf("unsupported column comparison: %s", function)
	}
}

// columnToValuePredicate builds a predicate comparing a column to a value.
func columnToValuePredicate(sel *sql.Selector, function, column string, value any) (*sql.Predicate, error) {
	col := sel.C(column)
	switch function {
	case filtering.FunctionEquals:
		return sql.EQ(col, value), nil
	case filtering.FunctionNotEquals:
		return sql.NEQ(col, value), nil
	case filtering.FunctionLessThan:
		return sql.LT(col, value), nil
	case filtering.FunctionLessEquals:
		return sql.LTE(col, value), nil
	case filtering.FunctionGreaterThan:
		return sql.GT(col, value), nil
	case filtering.FunctionGreaterEquals:
		return sql.GTE(col, value), nil
	default:
		return nil, fmt.Errorf("unsupported comparison operator %s", function)
	}
}

// valueToColumnPredicate builds a predicate comparing a value to a column.
func valueToColumnPredicate(sel *sql.Selector, function string, value any, column string) (*sql.Predicate, error) {
	col := sel.C(column)
	switch function {
	case filtering.FunctionEquals:
		return sql.EQ(col, value), nil
	case filtering.FunctionNotEquals:
		return sql.NEQ(col, value), nil
	case filtering.FunctionLessThan:
		return sql.GT(col, value), nil
	case filtering.FunctionLessEquals:
		return sql.GTE(col, value), nil
	case filtering.FunctionGreaterThan:
		return sql.LT(col, value), nil
	case filtering.FunctionGreaterEquals:
		return sql.LTE(col, value), nil
	default:
		return nil, fmt.Errorf("unsupported comparison operator %s", function)
	}
}

// exprToColumnExpr converts an expression to a column name string.
func exprToColumnExpr(e *expr.Expr) (string, bool) {
	switch kind := e.GetExprKind().(type) {
	case *expr.Expr_IdentExpr:
		return kind.IdentExpr.GetName(), true
	case *expr.Expr_SelectExpr:
		return exprToColumn(kind.SelectExpr)
	default:
		return "", false
	}
}

// exprToColumn converts a Select expression to a column name string.
func exprToColumn(sel *expr.Expr_Select) (string, bool) {
	if sel == nil {
		return "", false
	}
	if operand := sel.GetOperand(); operand != nil {
		if prefix, ok := exprToColumnExpr(operand); ok {
			return prefix + "." + sel.GetField(), true
		}
	}
	return sel.GetField(), sel.GetField() != ""
}

// exprToValue converts an expression to a Go value.
func exprToValue(e *expr.Expr) (any, error) {
	switch kind := e.GetExprKind().(type) {
	case *expr.Expr_ConstExpr:
		return constToValue(kind.ConstExpr)
	case *expr.Expr_CallExpr:
		call := kind.CallExpr
		switch call.GetFunction() {
		case filtering.FunctionTimestamp:
			if len(call.Args) != 1 {
				return nil, fmt.Errorf("timestamp expects 1 argument, got %d", len(call.Args))
			}
			argVal, err := exprToValue(call.Args[0])
			if err != nil {
				return nil, err
			}
			str, ok := argVal.(string)
			if !ok {
				return nil, fmt.Errorf("timestamp argument must be string, got %T", argVal)
			}
			tm, err := time.Parse(time.RFC3339, str)
			if err != nil {
				return nil, err
			}
			return tm, nil
		case filtering.FunctionDuration:
			if len(call.Args) != 1 {
				return nil, fmt.Errorf("duration expects 1 argument, got %d", len(call.Args))
			}
			argVal, err := exprToValue(call.Args[0])
			if err != nil {
				return nil, err
			}
			str, ok := argVal.(string)
			if !ok {
				return nil, fmt.Errorf("duration argument must be string, got %T", argVal)
			}
			d, err := time.ParseDuration(str)
			if err != nil {
				return nil, err
			}
			return d, nil
		default:
			return nil, fmt.Errorf("unsupported value function: %s", call.GetFunction())
		}
	default:
		return nil, fmt.Errorf("unsupported value expression %T", e.GetExprKind())
	}
}
