package filtering

import (
	"fmt"
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
	scope, ok := d.scopes[name]
	return scope, ok
}
