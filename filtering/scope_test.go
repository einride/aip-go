package filtering

import (
	"slices"
	"testing"

	freightv1 "go.einride.tech/aip/proto/gen/einride/example/freight/v1"
	syntaxv1 "go.einride.tech/aip/proto/gen/einride/example/syntax/v1"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
)

// scopeTestDeclarations declares shipment fields and a lane function at the root, and line item fields inside the
// scoped function line_items.any. Both shipments and line items have an external_reference_id field.
func scopeTestDeclarations(scopeOptions ...DeclarationOption) []DeclarationOption {
	return slices.Concat(
		DeclareProtoMessageIdents(
			&freightv1.Shipment{},
			WithFilterableFields("origin_site", "pickup_earliest_time", "external_reference_id"),
		),
		[]DeclarationOption{
			DeclareStandardFunctions(),
			DeclareFunction("lane", NewFunctionOverload("lane_string", TypeBool, TypeString)),
			DeclareScopedFunction("line_items.any", slices.Concat(
				DeclareProtoMessageIdents(
					&freightv1.LineItem{},
					WithFilterableFields("title", "weight_kg", "external_reference_id"),
				),
				[]DeclarationOption{DeclareStandardFunctions()},
				scopeOptions,
			)...),
		},
	)
}

func TestChecker_ScopedFunction(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		filter        string
		scopeOptions  []DeclarationOption
		errorContains string
	}{
		{
			name: "root and scope fields in their places",
			filter: `pickup_earliest_time > timestamp("2026-01-01T00:00:00Z") AND ` +
				`line_items.any((title = "pallet" AND weight_kg > 100.0))`,
		},
		{
			name:   "same name at the root and in the scope",
			filter: `external_reference_id = "s1" AND line_items.any((external_reference_id = "l1"))`,
		},
		{
			name:   "scope combined with root logic",
			filter: `NOT line_items.any((title = "pallet")) OR origin_site = "shippers/s1/sites/a"`,
		},
		{
			name:   "same scope called several times",
			filter: `line_items.any((title = "pallet")) AND line_items.any((weight_kg > 100.0))`,
		},
		{
			name:          "scope field at the root",
			filter:        `title = "pallet"`,
			errorContains: "undeclared identifier 'title'",
		},
		{
			name:          "root field inside the scope",
			filter:        `line_items.any((origin_site = "shippers/s1/sites/a"))`,
			errorContains: "check scoped function 'line_items.any' argument",
		},
		{
			name:          "root field inside the scope is undeclared",
			filter:        `line_items.any((origin_site = "shippers/s1/sites/a"))`,
			errorContains: "undeclared identifier 'origin_site'",
		},
		{
			name:          "root function inside the scope",
			filter:        `line_items.any((lane("a")))`,
			errorContains: "undeclared function 'lane'",
		},
		{
			name:          "scope inside the scope",
			filter:        `line_items.any((line_items.any((title = "pallet"))))`,
			errorContains: "undeclared function 'line_items.any'",
		},
		{
			name:   "scope declared inside the scope",
			filter: `line_items.any((line_items.any((title = "pallet"))))`,
			scopeOptions: []DeclarationOption{
				DeclareScopedFunction("line_items.any",
					DeclareStandardFunctions(),
					DeclareIdent("title", TypeString),
				),
			},
		},
		{
			name:          "non-bool argument",
			filter:        `line_items.any(title)`,
			errorContains: "non-bool result type",
		},
		{
			name:          "too many arguments",
			filter:        `line_items.any(title, title)`,
			errorContains: "scoped function 'line_items.any' takes exactly one argument",
		},
		{
			name:          "no arguments",
			filter:        `line_items.any()`,
			errorContains: "scoped function 'line_items.any' takes exactly one argument",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			declarations, err := NewDeclarations(scopeTestDeclarations(tt.scopeOptions...)...)
			assert.NilError(t, err)
			_, err = ParseFilterString(tt.filter, declarations)
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
				return
			}
			assert.NilError(t, err)
		})
	}
}

func TestChecker_ScopedFunction_StandardFunctionsNotInherited(t *testing.T) {
	t.Parallel()
	declarations, err := NewDeclarations(
		DeclareStandardFunctions(),
		DeclareScopedFunction("line_items.any", DeclareIdent("title", TypeString)),
	)
	assert.NilError(t, err)
	_, err = ParseFilterString(`line_items.any((title = "pallet"))`, declarations)
	assert.ErrorContains(t, err, "undeclared function '='")
}

// TestChecker_ScopedFunction_TypeMap verifies that a name declared at the root and inside a scope with different
// types is typed by where it is used, that enum values declared at the root are not visible inside the scope, and
// that every expression, inside the scope too, is in the type map.
func TestChecker_ScopedFunction_TypeMap(t *testing.T) {
	t.Parallel()
	declarations, err := NewDeclarations(
		DeclareStandardFunctions(),
		DeclareEnumIdent("enum", syntaxv1.Enum(0).Type()),
		DeclareScopedFunction("items.any",
			DeclareStandardFunctions(),
			DeclareIdent("enum", TypeString),
		),
	)
	assert.NilError(t, err)
	_, err = ParseFilterString(`items.any((enum = ENUM_ONE))`, declarations)
	assert.ErrorContains(t, err, "undeclared identifier 'ENUM_ONE'")

	filter, err := ParseFilterString(`enum = ENUM_ONE AND items.any((enum = "one"))`, declarations)
	assert.NilError(t, err)
	typeMap := filter.CheckedExpr.GetTypeMap()
	root := filter.CheckedExpr.GetExpr().GetCallExpr()
	rootEnum := root.GetArgs()[0].GetCallExpr().GetArgs()[0]
	scopeCall := root.GetArgs()[1]
	scopeEnum := scopeCall.GetCallExpr().GetArgs()[0].GetCallExpr().GetArgs()[0]
	assert.Equal(t, "enum", rootEnum.GetIdentExpr().GetName())
	assert.Equal(t, "enum", scopeEnum.GetIdentExpr().GetName())
	assert.DeepEqual(t, TypeEnum(syntaxv1.Enum(0).Type()), typeMap[rootEnum.GetId()], protocmp.Transform())
	assert.DeepEqual(t, TypeString, typeMap[scopeEnum.GetId()], protocmp.Transform())
	assert.DeepEqual(t, TypeBool, typeMap[scopeCall.GetId()], protocmp.Transform())
	Walk(func(currExpr, _ *expr.Expr) bool {
		_, ok := typeMap[currExpr.GetId()]
		assert.Assert(t, ok, "missing type of expression %v", currExpr)
		return true
	}, filter.CheckedExpr.GetExpr())
}

// TestChecker_ScopedFunction_TypeMapSelectPath verifies that a declared select path inside a scope is typed as at the
// root: the path has a type and its operands have none, rather than a nil type.
func TestChecker_ScopedFunction_TypeMapSelectPath(t *testing.T) {
	t.Parallel()
	declarations, err := NewDeclarations(
		DeclareStandardFunctions(),
		DeclareScopedFunction("items.any",
			DeclareStandardFunctions(),
			DeclareIdent("pickup.site", TypeString),
		),
	)
	assert.NilError(t, err)
	filter, err := ParseFilterString(`items.any((pickup.site = "a"))`, declarations)
	assert.NilError(t, err)
	typeMap := filter.CheckedExpr.GetTypeMap()
	path := filter.CheckedExpr.GetExpr().GetCallExpr().GetArgs()[0].GetCallExpr().GetArgs()[0]
	assert.DeepEqual(t, TypeString, typeMap[path.GetId()], protocmp.Transform())
	_, ok := typeMap[path.GetSelectExpr().GetOperand().GetId()]
	assert.Assert(t, !ok, "operand of a declared select path has a type")
	for id, exprType := range typeMap {
		assert.Assert(t, exprType != nil, "nil type for expression %d", id)
	}
}

func TestDeclareScopedFunction_Redeclaration(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		declarations  []DeclarationOption
		errorContains string
	}{
		{
			name: "scoped function twice",
			declarations: []DeclarationOption{
				DeclareScopedFunction("line_items.any"),
				DeclareScopedFunction("line_items.any"),
			},
			errorContains: "redeclaration of line_items.any",
		},
		{
			name: "scoped function after function",
			declarations: []DeclarationOption{
				DeclareFunction("line_items.any", NewFunctionOverload("line_items_any_bool", TypeBool, TypeBool)),
				DeclareScopedFunction("line_items.any"),
			},
			errorContains: "redeclaration of line_items.any",
		},
		{
			name: "function after scoped function",
			declarations: []DeclarationOption{
				DeclareScopedFunction("line_items.any"),
				DeclareFunction("line_items.any", NewFunctionOverload("line_items_any_bool", TypeBool, TypeBool)),
			},
			errorContains: "redeclaration of line_items.any",
		},
		{
			name: "invalid scope declarations",
			declarations: []DeclarationOption{
				DeclareScopedFunction("line_items.any",
					DeclareIdent("title", TypeString),
					DeclareIdent("title", TypeInt),
				),
			},
			errorContains: "declare scoped function line_items.any: redeclaration of title",
		},
		{
			name: "same name with different types in root and scope",
			declarations: []DeclarationOption{
				DeclareIdent("title", TypeInt),
				DeclareScopedFunction("line_items.any", DeclareIdent("title", TypeString)),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewDeclarations(tt.declarations...)
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
				return
			}
			assert.NilError(t, err)
		})
	}
}
