package filtering

import (
	"testing"

	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
)

func TestApplyMacros(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name              string
		filter            string
		declarations      []DeclarationOption
		macros            []Macro
		macroDeclarations []DeclarationOption
		expected          *expr.Expr
		errorContains     string
	}{
		{
			name:   `annotations.schedule = "test" --> annotations: "schedule=test"`,
			filter: `annotations.schedule = "test"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("annotations", TypeMap(TypeString, TypeString)),
			},
			macros: []Macro{
				func(cursor *Cursor) {
					callExpr := cursor.Expr().GetCallExpr()
					if callExpr == nil {
						return
					}
					if callExpr.GetFunction() != FunctionEquals {
						return
					}
					if len(callExpr.GetArgs()) != 2 {
						return
					}
					arg0Select := callExpr.GetArgs()[0].GetSelectExpr()
					if arg0Select == nil || arg0Select.GetOperand().GetIdentExpr().GetName() != "annotations" {
						return
					}
					arg1String := callExpr.GetArgs()[1].GetConstExpr().GetStringValue()
					if arg1String == "" {
						return
					}
					cursor.Replace(Has(arg0Select.GetOperand(), String(arg0Select.GetField()+"="+arg1String)))
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("annotations", TypeList(TypeString)),
			},
			expected: Has(Text("annotations"), String("schedule=test")),
		},
		{
			name:   "no-op macro that doesn't match",
			filter: `name = "test"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name", TypeString),
			},
			macros: []Macro{
				func(cursor *Cursor) {
					// Macro that only matches "other_name", so this should be a no-op
					identExpr := cursor.Expr().GetIdentExpr()
					if identExpr == nil || identExpr.GetName() != "other_name" {
						return
					}
					cursor.Replace(Text("renamed"))
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name", TypeString),
			},
			expected: Equals(Text("name"), String("test")),
		},
		{
			name:   "empty macros list is no-op",
			filter: `name = "test"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name", TypeString),
			},
			macros: []Macro{},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name", TypeString),
			},
			expected: Equals(Text("name"), String("test")),
		},
		{
			name:   "multiple macros applied in sequence",
			filter: `x = 5 AND y = 10`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("x", TypeInt),
				DeclareIdent("y", TypeInt),
			},
			macros: []Macro{
				// First macro: rename x to x_renamed
				func(cursor *Cursor) {
					identExpr := cursor.Expr().GetIdentExpr()
					if identExpr == nil || identExpr.GetName() != "x" {
						return
					}
					cursor.Replace(Text("x_renamed"))
				},
				// Second macro: rename y to y_renamed
				func(cursor *Cursor) {
					identExpr := cursor.Expr().GetIdentExpr()
					if identExpr == nil || identExpr.GetName() != "y" {
						return
					}
					cursor.Replace(Text("y_renamed"))
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("x_renamed", TypeInt),
				DeclareIdent("y_renamed", TypeInt),
			},
			expected: And(
				Equals(Text("x_renamed"), Int(5)),
				Equals(Text("y_renamed"), Int(10)),
			),
		},
		{
			name:   "macro on nested expression",
			filter: `(x = 5) AND (y = 10)`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("x", TypeInt),
				DeclareIdent("y", TypeInt),
			},
			macros: []Macro{
				func(cursor *Cursor) {
					callExpr := cursor.Expr().GetCallExpr()
					if callExpr == nil || callExpr.GetFunction() != FunctionEquals {
						return
					}
					if len(callExpr.GetArgs()) != 2 {
						return
					}
					arg0Ident := callExpr.GetArgs()[0].GetIdentExpr()
					if arg0Ident == nil || arg0Ident.GetName() != "x" {
						return
					}
					// Transform x = 5 to x_renamed = 5
					cursor.Replace(Equals(Text("x_renamed"), callExpr.GetArgs()[1]))
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("x_renamed", TypeInt),
				DeclareIdent("y", TypeInt),
			},
			expected: And(
				Equals(Text("x_renamed"), Int(5)),
				Equals(Text("y"), Int(10)),
			),
		},
		{
			name:   "same ident appears multiple times",
			filter: `name = "test" AND name = "test2"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name", TypeString),
			},
			macros: []Macro{
				func(cursor *Cursor) {
					identExpr := cursor.Expr().GetIdentExpr()
					if identExpr == nil || identExpr.GetName() != "name" {
						return
					}
					cursor.Replace(Text("name_renamed"))
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name_renamed", TypeString),
			},
			expected: And(
				Equals(Text("name_renamed"), String("test")),
				Equals(Text("name_renamed"), String("test2")),
			),
		},
		{
			name:   "macro using ReplaceWithDeclarations with same declaration done as validation",
			filter: `name = "test"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name", TypeString),
			},
			macros: []Macro{
				func(cursor *Cursor) {
					identExpr := cursor.Expr().GetIdentExpr()
					if identExpr == nil || identExpr.GetName() != "name" {
						return
					}
					cursor.ReplaceWithDeclarations(
						Text("renamed_name"),
						[]DeclarationOption{
							DeclareIdent("renamed_name", TypeString),
						},
					)
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("renamed_name", TypeString),
			},
			expected: Equals(Text("renamed_name"), String("test")),
		},
		{
			name:   "macro using ReplaceWithDeclarations, declaration in macro only",
			filter: `name = "test"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name", TypeString),
			},
			macros: []Macro{
				func(cursor *Cursor) {
					identExpr := cursor.Expr().GetIdentExpr()
					if identExpr == nil || identExpr.GetName() != "name" {
						return
					}
					cursor.ReplaceWithDeclarations(
						Text("renamed_name"),
						[]DeclarationOption{
							DeclareIdent("renamed_name", TypeString),
						},
					)
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				// No declaration for renamed_name here
			},
			errorContains: "undeclared identifier 'renamed_name'",
		},
		{
			name:   `name != "test" --> NOT(name = "test")`,
			filter: `name != "test"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name", TypeString),
			},
			macros: []Macro{
				func(cursor *Cursor) {
					callExpr := cursor.Expr().GetCallExpr()
					if callExpr == nil || callExpr.GetFunction() != FunctionNotEquals {
						return
					}
					if len(callExpr.GetArgs()) != 2 {
						return
					}
					// Transform name != "test" to NOT(name = "test")
					cursor.Replace(Not(Equals(callExpr.GetArgs()[0], callExpr.GetArgs()[1])))
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name", TypeString),
			},
			expected: Not(Equals(Text("name"), String("test"))),
		},
		{
			name:   `age < 18 --> age <= 17`,
			filter: `age < 18`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("age", TypeInt),
			},
			macros: []Macro{
				func(cursor *Cursor) {
					callExpr := cursor.Expr().GetCallExpr()
					if callExpr == nil || callExpr.GetFunction() != FunctionLessThan {
						return
					}
					if len(callExpr.GetArgs()) != 2 {
						return
					}
					// Transform age < 18 to age <= 17
					arg0 := callExpr.GetArgs()[0]
					arg1Int := callExpr.GetArgs()[1].GetConstExpr().GetInt64Value()
					cursor.Replace(LessEquals(arg0, Int(arg1Int-1)))
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("age", TypeInt),
			},
			expected: LessEquals(Text("age"), Int(17)),
		},
		{
			name:   `x = 1 OR y = 2 --> NOT(NOT(x = 1) AND NOT(y = 2))`,
			filter: `x = 1 OR y = 2`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("x", TypeInt),
				DeclareIdent("y", TypeInt),
			},
			macros: []Macro{
				func(cursor *Cursor) {
					callExpr := cursor.Expr().GetCallExpr()
					if callExpr == nil || callExpr.GetFunction() != FunctionOr {
						return
					}
					if len(callExpr.GetArgs()) != 2 {
						return
					}
					// Transform OR to NOT(AND(NOT(...), NOT(...)))
					cursor.Replace(Not(And(Not(callExpr.GetArgs()[0]), Not(callExpr.GetArgs()[1]))))
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("x", TypeInt),
				DeclareIdent("y", TypeInt),
			},
			expected: Not(And(
				Not(Equals(Text("x"), Int(1))),
				Not(Equals(Text("y"), Int(2))),
			)),
		},
		{
			name:   `user.name = "John" --> user_name = "John"`,
			filter: `user.name = "John"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("user", TypeMap(TypeString, TypeString)),
			},
			macros: []Macro{
				func(cursor *Cursor) {
					selectExpr := cursor.Expr().GetSelectExpr()
					if selectExpr == nil {
						return
					}
					operandIdent := selectExpr.GetOperand().GetIdentExpr()
					if operandIdent == nil || operandIdent.GetName() != "user" {
						return
					}
					if selectExpr.GetField() != "name" {
						return
					}
					// Transform user.name to user_name
					cursor.Replace(Text("user_name"))
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("user_name", TypeString),
			},
			expected: Equals(Text("user_name"), String("John")),
		},
		{
			name:   "macro stops traversal after replacement",
			filter: `x = 5 AND y = 10`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("x", TypeInt),
				DeclareIdent("y", TypeInt),
			},
			macros: []Macro{
				// Macro that replaces the entire AND expression
				func(cursor *Cursor) {
					callExpr := cursor.Expr().GetCallExpr()
					if callExpr == nil || callExpr.GetFunction() != FunctionAnd {
						return
					}
					// Replace entire AND with a single condition
					cursor.Replace(Equals(Text("combined"), Int(15)))
				},
				// This macro should not run because the parent was replaced
				func(cursor *Cursor) {
					identExpr := cursor.Expr().GetIdentExpr()
					if identExpr == nil || identExpr.GetName() == "x" {
						cursor.Replace(Text("x_should_not_match"))
					}
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("combined", TypeInt),
			},
			expected: Equals(Text("combined"), Int(15)),
		},
		{
			name:   "type checking error after macro application",
			filter: `name = "test"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name", TypeString),
			},
			macros: []Macro{
				func(cursor *Cursor) {
					identExpr := cursor.Expr().GetIdentExpr()
					if identExpr == nil || identExpr.GetName() != "name" {
						return
					}
					// Replace with int type, but filter expects string - should cause type error
					cursor.Replace(Text("name"))
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name", TypeInt), // Wrong type - should cause error
			},
			errorContains: "no matching overload",
		},
		{
			name:   "complex nested structure with multiple macro matches",
			filter: `(a = 1 AND b = 2) OR (c = 3 AND d = 4)`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("a", TypeInt),
				DeclareIdent("b", TypeInt),
				DeclareIdent("c", TypeInt),
				DeclareIdent("d", TypeInt),
			},
			macros: []Macro{
				// Macro that adds 10 to all integer constants
				func(cursor *Cursor) {
					constExpr := cursor.Expr().GetConstExpr()
					if constExpr == nil {
						return
					}
					if _, ok := constExpr.GetConstantKind().(*expr.Constant_Int64Value); ok {
						intVal := constExpr.GetInt64Value()
						cursor.Replace(Int(intVal + 10))
					}
				},
			},
			macroDeclarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("a", TypeInt),
				DeclareIdent("b", TypeInt),
				DeclareIdent("c", TypeInt),
				DeclareIdent("d", TypeInt),
			},
			expected: Or(
				And(
					Equals(Text("a"), Int(11)),
					Equals(Text("b"), Int(12)),
				),
				And(
					Equals(Text("c"), Int(13)),
					Equals(Text("d"), Int(14)),
				),
			),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			declarations, err := NewDeclarations(tt.declarations...)
			assert.NilError(t, err)
			filter, err := ParseFilter(&mockRequest{filter: tt.filter}, declarations)
			if err != nil && tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
				return
			}
			assert.NilError(t, err)
			macroDeclarations, err := NewDeclarations(tt.macroDeclarations...)
			assert.NilError(t, err)
			actual, err := ApplyMacros(filter, macroDeclarations, tt.macros...)
			if err != nil && tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(
				t,
				tt.expected,
				actual.CheckedExpr.GetExpr(),
				protocmp.Transform(),
				protocmp.IgnoreFields(&expr.Expr{}, "id"),
			)
		})
	}
}

type mockRequest struct {
	filter string
}

func (m *mockRequest) GetFilter() string {
	return m.filter
}
