package filtering

import (
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"
)

// Filter represents a parsed and type-checked filter.
type Filter struct {
	CheckedExpr  *expr.CheckedExpr
	declarations *Declarations
}

// WithMacros returns a new Filter with the given macros applied and the
// result type-checked. f is not modified.
//
// It is safe to call WithMacros concurrently on filters that share the same
// underlying *Declarations, and safe to call repeatedly on the same Filter
// (e.g. when retrying a Spanner transaction): f.CheckedExpr is cloned before
// rewriting, so f retains its original expression tree on every invocation.
//
// EXPERIMENTAL: This method is experimental and may be changed or removed in the future.
func (f Filter) WithMacros(macros ...Macro) (Filter, error) {
	// Clone the CheckedExpr so the macro rewrite does not mutate f's
	// expression tree.
	rewritten := proto.CloneOf(f.CheckedExpr)
	declarationOptions, err := applyMacros(
		rewritten.GetExpr(),
		rewritten.GetSourceInfo(),
		f.declarations,
		macros...,
	)
	if err != nil {
		return Filter{}, err
	}
	newDeclarations, err := NewDeclarations(declarationOptions...)
	if err != nil {
		return Filter{}, err
	}
	// Clone declarations so f.declarations is not mutated.
	declarations := f.declarations.clone()
	declarations.merge(newDeclarations)
	var checker Checker
	checker.Init(rewritten.GetExpr(), rewritten.GetSourceInfo(), declarations)
	checkedExpr, err := checker.Check()
	if err != nil {
		return Filter{}, err
	}
	return Filter{
		CheckedExpr:  checkedExpr,
		declarations: declarations,
	}, nil
}
