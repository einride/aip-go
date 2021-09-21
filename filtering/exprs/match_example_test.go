package exprs

import (
	"fmt"

	"go.einride.tech/aip/filtering"
	"go.einride.tech/aip/resourcename"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func ExampleMatchFunction_validateResourceNames() {
	const pattern = "books/{book}"

	var walkErr error
	walkFn := func(currExpr, _ *expr.Expr) bool {
		var name string
		// match any function expression with name '=' and LHS 'name'
		matcher := MatchFunction(filtering.FunctionEquals, MatchText("name"), MatchAnyString(&name))
		if matcher(currExpr) && !resourcename.Match(pattern, name) {
			walkErr = fmt.Errorf("expected resource name matching '%s' but got '%s'", pattern, name)
			return false
		}
		return true
	}

	// name = "not a resource name" or name = "books/2"
	invalidExpr := filtering.Or(
		filtering.Equals(filtering.Text("name"), filtering.String("not a resource name")),
		filtering.Equals(filtering.Text("name"), filtering.String("books/2")),
	)
	filtering.Walk(walkFn, invalidExpr)
	fmt.Println(walkErr)

	// reset
	walkErr = nil

	// name = "books/1" or name = "books/2"
	validExpr := filtering.Or(
		filtering.Equals(filtering.Text("name"), filtering.String("books/1")),
		filtering.Equals(filtering.Text("name"), filtering.String("books/2")),
	)
	filtering.Walk(walkFn, validExpr)
	fmt.Println(walkErr)

	// Output:
	// expected resource name matching 'books/{book}' but got 'not a resource name'
	// <nil>
}
