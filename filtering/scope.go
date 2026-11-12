package filtering

import (
	"fmt"

	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"
)

// DeclareScopedFunction declares a function that takes one boolean argument and returns a boolean. The argument is
// type-checked against options only: none of the enclosing declarations are visible inside it, the standard
// functions included, unless options declare them.
//
// A scoped function lets a filter express conditions in another context than the filtered resource, such as on the
// elements of a repeated field. For example, with
//
//	DeclareScopedFunction("items.any", DeclareStandardFunctions(), DeclareIdent("price", TypeInt))
//
// the filter `items.any((price > 100))` is valid, while `price > 100` outside the scoped function is an undeclared
// identifier. The same name can be declared both outside and inside a scoped function, with different types, and is
// resolved by where it is used.
//
// EXPERIMENTAL: This function is experimental and may be changed or removed in the future.
func DeclareScopedFunction(name string, options ...DeclarationOption) DeclarationOption {
	return func(declarations *Declarations) error {
		scope, err := NewDeclarations(options...)
		if err != nil {
			return fmt.Errorf("declare scoped function %s: %w", name, err)
		}
		return declarations.declareScopedFunction(name, scope)
	}
}

func (d *Declarations) declareScopedFunction(name string, scope *Declarations) error {
	if _, ok := d.scopes[name]; ok {
		return fmt.Errorf("redeclaration of %s", name)
	}
	if _, ok := d.functions[name]; ok {
		return fmt.Errorf("redeclaration of %s", name)
	}
	d.scopes[name] = scope
	return nil
}

// lookupScopedFunction returns the declarations that the argument of the scoped function name is checked against.
func (d *Declarations) lookupScopedFunction(name string) (*Declarations, bool) {
	if d == nil {
		return nil, false
	}
	scope, ok := d.scopes[name]
	return scope, ok
}

// scopedCall returns the name of the scoped function that e calls, if e is a call to one of the scoped functions in
// declarations.
func scopedCall(e *expr.Expr, declarations *Declarations) (string, bool) {
	function := e.GetCallExpr().GetFunction()
	if _, ok := declarations.lookupScopedFunction(function); !ok {
		return "", false
	}
	return function, true
}

// WithScopeMacros returns a new Filter with the given macros applied inside the arguments of calls to the scoped
// function name, and the result type-checked. Declarations that macros add with Cursor.ReplaceWithDeclarations or
// Cursor.AddDeclarations are added to the scope, not to the root. f is not modified.
//
// Filter.WithMacros doesn't apply macros inside the arguments of scoped functions, so macros for the root and for
// each scope are applied separately:
//
//	filter, err := filter.WithMacros(rootMacros...)
//	// ...
//	filter, err = filter.WithScopeMacros("items.any", itemMacros...)
//
// Only calls to name at the root are rewritten, not calls in other scopes. An empty filter is returned unchanged.
//
// EXPERIMENTAL: This method is experimental and may be changed or removed in the future.
func (f Filter) WithScopeMacros(name string, macros ...Macro) (Filter, error) {
	if f.CheckedExpr == nil {
		return f, nil
	}
	scope, ok := f.declarations.lookupScopedFunction(name)
	if !ok {
		return Filter{}, fmt.Errorf("undeclared scoped function '%s'", name)
	}
	// Clone the CheckedExpr so the macro rewrite does not mutate f's expression tree.
	rewritten := proto.CloneOf(f.CheckedExpr)
	declarationOptions, err := applyMacrosInScopedCalls(
		rewritten.GetExpr(), rewritten.GetSourceInfo(), f.declarations, name, scope, macros...,
	)
	if err != nil {
		return Filter{}, fmt.Errorf("apply macros to scoped function '%s': %w", name, err)
	}
	declarations, err := f.declarations.withScopeDeclarations(name, declarationOptions...)
	if err != nil {
		return Filter{}, fmt.Errorf("declare macro output in scoped function '%s': %w", name, err)
	}
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

// applyMacrosInScopedCalls applies macros inside the arguments of the calls to the scoped function name in root, and
// returns the declaration options that the macros added. declarations are those of root's scope, and scope those of
// the arguments. Calls to name inside other scoped functions' arguments are in other scopes, and are not rewritten.
func applyMacrosInScopedCalls(
	root *expr.Expr,
	sourceInfo *expr.SourceInfo,
	declarations *Declarations,
	name string,
	scope *Declarations,
	macros ...Macro,
) ([]DeclarationOption, error) {
	var declarationOptions []DeclarationOption
	var applyErr error
	Walk(func(currExpr, _ *expr.Expr) bool {
		if applyErr != nil {
			return false
		}
		function, ok := scopedCall(currExpr, declarations)
		if !ok {
			return true
		}
		if function != name {
			return false
		}
		for _, arg := range currExpr.GetCallExpr().GetArgs() {
			// The root, not the argument, gives the next free expression ID, as the root also contains the expressions
			// that macros added to earlier arguments.
			argDeclarationOptions, err := applyMacrosWithRoot(root, arg, sourceInfo, scope, macros...)
			if err != nil {
				applyErr = err
				return false
			}
			declarationOptions = append(declarationOptions, argDeclarationOptions...)
		}
		return false
	}, root)
	if applyErr != nil {
		return nil, applyErr
	}
	return declarationOptions, nil
}

// withScopeDeclarations returns a copy of d where the declarations of the scoped function name also have the given
// options declared. Neither d nor the scope's declarations are modified.
func (d *Declarations) withScopeDeclarations(name string, options ...DeclarationOption) (*Declarations, error) {
	addedDeclarations, err := NewDeclarations(options...)
	if err != nil {
		return nil, err
	}
	scope := d.scopes[name].clone()
	scope.merge(addedDeclarations)
	result := d.clone()
	result.scopes[name] = scope
	return result, nil
}
