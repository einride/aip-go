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
	} {
		tt := tt
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
