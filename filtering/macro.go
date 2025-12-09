package filtering

import (
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// Macro represents a function that can perform macro replacements on a filter expression.
type Macro func(*Cursor)

// ApplyMacros applies the provided macros to the filter and type-checks the result against the provided declarations.
func ApplyMacros(filter Filter, declarations *Declarations, macros ...Macro) (Filter, error) {
	// We ignore the return value as we validate against the given declarations instead.
	_ = applyMacros(filter.CheckedExpr.GetExpr(), filter.CheckedExpr.GetSourceInfo(), filter.declarations, macros...)
	var checker Checker
	checker.Init(filter.CheckedExpr.GetExpr(), filter.CheckedExpr.GetSourceInfo(), declarations)
	checkedExpr, err := checker.Check()
	if err != nil {
		return Filter{}, err
	}
	filter.CheckedExpr = checkedExpr
	return filter, nil
}

func applyMacros(
	exp *expr.Expr,
	sourceInfo *expr.SourceInfo,
	declarations *Declarations,
	macros ...Macro,
) []DeclarationOption {
	declarationOptions := make([]DeclarationOption, 0, len(macros))
	nextID := maxID(exp) + 1
	Walk(func(currExpr, parentExpr *expr.Expr) bool {
		cursor := &Cursor{
			sourceInfo:       sourceInfo,
			currExpr:         currExpr,
			parentExpr:       parentExpr,
			nextID:           nextID,
			exprDeclarations: declarations,
		}
		for _, macro := range macros {
			macro(cursor)
			nextID = cursor.nextID
			declarationOptions = append(declarationOptions, cursor.replaceDeclOptions...)
			if cursor.replaced {
				// Don't traverse children of replaced expr.
				return false
			}
		}
		return true
	}, exp)
	return declarationOptions
}

// A Cursor describes an expression encountered while applying a Macro.
//
// The method Replace can be used to rewrite the filter.
type Cursor struct {
	parentExpr         *expr.Expr
	currExpr           *expr.Expr
	sourceInfo         *expr.SourceInfo
	exprDeclarations   *Declarations
	replaced           bool
	nextID             int64
	replaceDeclOptions []DeclarationOption
}

// Parent returns the parent of the current expression.
func (c *Cursor) Parent() (*expr.Expr, bool) {
	return c.parentExpr, c.parentExpr != nil
}

// Expr returns the current expression.
func (c *Cursor) Expr() *expr.Expr {
	return c.currExpr
}

// LookupIdentType looks up the type of an ident in the filter declarations.
// EXPERIMENTAL: This method is experimental and may be changed or removed in the future.
func (c *Cursor) LookupIdentType(name string) (*expr.Type, bool) {
	ident, ok := c.exprDeclarations.LookupIdent(name)
	if !ok {
		return nil, false
	}
	return ident.GetIdent().GetType(), true
}

// Replace the current expression with a new expression.
func (c *Cursor) Replace(newExpr *expr.Expr) {
	Walk(func(childExpr, _ *expr.Expr) bool {
		childExpr.Id = c.nextID
		c.nextID++
		return true
	}, newExpr)
	if c.sourceInfo.MacroCalls == nil {
		c.sourceInfo.MacroCalls = map[int64]*expr.Expr{}
	}
	c.sourceInfo.MacroCalls[newExpr.GetId()] = &expr.Expr{Id: c.currExpr.GetId(), ExprKind: c.currExpr.GetExprKind()}
	c.currExpr.Id = newExpr.GetId()
	c.currExpr.ExprKind = newExpr.GetExprKind()
	c.replaced = true
}

// ReplaceWithType replaces the current expression with a new  expression and type.
// EXPERIMENTAL: This method is experimental and may be changed or removed in the future.
func (c *Cursor) ReplaceWithDeclarations(newExpr *expr.Expr, opts []DeclarationOption) {
	Walk(func(childExpr, _ *expr.Expr) bool {
		childExpr.Id = c.nextID
		c.nextID++
		return true
	}, newExpr)
	if c.sourceInfo.MacroCalls == nil {
		c.sourceInfo.MacroCalls = map[int64]*expr.Expr{}
	}
	c.sourceInfo.MacroCalls[newExpr.GetId()] = &expr.Expr{Id: c.currExpr.GetId(), ExprKind: c.currExpr.GetExprKind()}
	c.currExpr.Id = newExpr.GetId()
	c.currExpr.ExprKind = newExpr.GetExprKind()
	c.replaceDeclOptions = append(c.replaceDeclOptions, opts...)
	c.replaced = true
}

func maxID(exp *expr.Expr) int64 {
	var maxFound int64
	Walk(func(_, _ *expr.Expr) bool {
		if exp.GetId() > maxFound {
			maxFound = exp.GetId()
		}
		return true
	}, exp)
	return maxFound
}
