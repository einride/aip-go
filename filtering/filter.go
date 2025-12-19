package filtering

import (
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// Filter represents a parsed and type-checked filter.
type Filter struct {
	CheckedExpr  *expr.CheckedExpr
	declarations *Declarations
}

// ApplyMacros modifies the filter by applying the provided macros.
// EXPERIMENTAL: This method is experimental and may be changed or removed in the future.
func (f *Filter) ApplyMacros(macros ...Macro) error {
	declarationOptions, err := applyMacros(
		f.CheckedExpr.GetExpr(),
		f.CheckedExpr.GetSourceInfo(),
		f.declarations,
		macros...,
	)
	if err != nil {
		return err
	}
	newDeclarations, err := NewDeclarations(declarationOptions...)
	if err != nil {
		return err
	}
	f.declarations.merge(newDeclarations)
	var checker Checker
	checker.Init(f.CheckedExpr.GetExpr(), f.CheckedExpr.GetSourceInfo(), f.declarations)
	checkedExpr, err := checker.Check()
	if err != nil {
		return err
	}
	f.CheckedExpr = checkedExpr
	return nil
}
