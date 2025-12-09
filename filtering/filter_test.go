package filtering

import (
	"testing"

	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
)

func TestFilter_ApplyMacros(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		filter        string
		declarations  []DeclarationOption
		macros        []Macro
		expected      *expr.Expr
		errorContains string
	}{
		{
			name:   "identifier rename macro",
			filter: `name = "value"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name", TypeString),
			},
			macros: []Macro{
				func(cursor *Cursor) {
					identExpr := cursor.Expr().GetIdentExpr()
					if identExpr == nil {
						return
					}
					if identExpr.GetName() != "name" {
						return
					}
					cursor.ReplaceWithDeclarations(
						Text("renamed_name"),
						[]DeclarationOption{
							DeclareIdent("renamed_name", TypeString),
						})
				},
			},
			expected: Equals(Text("renamed_name"), String("value")),
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
					cursor.ReplaceWithDeclarations(Text("x_renamed"), []DeclarationOption{
						DeclareIdent("x_renamed", TypeInt),
					})
				},
				// Second macro: rename y to y_renamed
				func(cursor *Cursor) {
					identExpr := cursor.Expr().GetIdentExpr()
					if identExpr == nil || identExpr.GetName() != "y" {
						return
					}
					cursor.ReplaceWithDeclarations(Text("y_renamed"), []DeclarationOption{
						DeclareIdent("y_renamed", TypeInt),
					})
				},
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
					cursor.ReplaceWithDeclarations(
						Equals(Text("x_renamed"), callExpr.GetArgs()[1]),
						[]DeclarationOption{
							DeclareIdent("x_renamed", TypeInt),
						},
					)
				},
			},
			expected: And(
				Equals(Text("x_renamed"), Int(5)),
				Equals(Text("y"), Int(10)),
			),
		},
		{
			name:   "change equal function lhs from string to int",
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
						Text("name"),
						[]DeclarationOption{
							DeclareIdent("name", TypeInt),
							DeclareFunction(FunctionEquals, NewFunctionOverload(FunctionEquals, TypeBool, TypeInt, TypeString)),
						},
					)
				},
			},
		},
		{
			name:   "empty macros list is no-op",
			filter: `name = "test"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("name", TypeString),
			},
			macros:   []Macro{},
			expected: Equals(Text("name"), String("test")),
		},
		{
			name:   "same ident appears several times in the filter",
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
					cursor.ReplaceWithDeclarations(
						Text("name_renamed"),
						[]DeclarationOption{
							DeclareIdent("name_renamed", TypeString),
						},
					)
				},
			},
			expected: And(
				Equals(Text("name_renamed"), String("test")),
				Equals(Text("name_renamed"), String("test2")),
			),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			declarations, err := NewDeclarations(tt.declarations...)
			assert.NilError(t, err)
			filter, err := ParseFilter(&mockRequest{filter: tt.filter}, declarations)
			assert.NilError(t, err)

			err = filter.ApplyMacros(tt.macros...)
			if err != nil && tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
				return
			}
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
				return
			}
			assert.NilError(t, err)

			// Verify filter was modified in place
			if tt.expected != nil {
				assert.DeepEqual(
					t,
					tt.expected,
					filter.CheckedExpr.GetExpr(),
					protocmp.Transform(),
					protocmp.IgnoreFields(&expr.Expr{}, "id"),
				)
			}
		})
	}
}
