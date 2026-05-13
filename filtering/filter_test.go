package filtering

import (
	"sync"
	"testing"

	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
)

func TestFilter_WithMacros(t *testing.T) {
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
		{
			name:   "AddDeclarations without Replace",
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
					// Add an extra declaration without replacing the expression.
					cursor.AddDeclarations(DeclareIdent("extra", TypeString))
				},
			},
			// Expression is unchanged.
			expected: Equals(Text("name"), String("test")),
		},
		{
			name:   "float field equals int literal",
			filter: `foo = 3`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("foo", TypeFloat),
			},
			macros:   []Macro{},
			expected: Equals(Text("foo"), Int(3)),
		},
		{
			name:   "float field less than or equal to int literal",
			filter: `foo <= 3`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("foo", TypeFloat),
			},
			macros:   []Macro{},
			expected: LessEquals(Text("foo"), Int(3)),
		},
		{
			name:   "float field greater than or equals to int literal",
			filter: `foo >= 3`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("foo", TypeFloat),
			},
			macros:   []Macro{},
			expected: GreaterEquals(Text("foo"), Int(3)),
		},
		{
			name:   "float field less than int literal",
			filter: `foo < 3`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("foo", TypeFloat),
			},
			macros:   []Macro{},
			expected: LessThan(Text("foo"), Int(3)),
		},
		{
			name:   "float field greater than int literal",
			filter: `foo > 3`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("foo", TypeFloat),
			},
			macros:   []Macro{},
			expected: GreaterThan(Text("foo"), Int(3)),
		},
		{
			name:   "rename function call and inject declaration via AddDeclarations",
			filter: `fuzzySearch("hello")`,
			declarations: []DeclarationOption{
				DeclareFunction(
					"fuzzySearch",
					NewFunctionOverload("fuzzySearch", TypeBool, TypeString),
				),
			},
			macros: []Macro{
				func(cursor *Cursor) {
					callExpr := cursor.Expr().GetCallExpr()
					if callExpr == nil || callExpr.GetFunction() != "fuzzySearch" {
						return
					}
					// Rename to a storage-specific function and inject its declaration.
					cursor.Replace(Function("spannerSearch", callExpr.GetArgs()...))
					cursor.AddDeclarations(
						DeclareFunction(
							"spannerSearch",
							NewFunctionOverload("spannerSearch", TypeBool, TypeString),
						),
					)
				},
			},
			expected: Function("spannerSearch", String("hello")),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			declarations, err := NewDeclarations(tt.declarations...)
			assert.NilError(t, err)
			source, err := ParseFilter(&mockRequest{filter: tt.filter}, declarations)
			assert.NilError(t, err)

			result, err := source.WithMacros(tt.macros...)
			if err != nil && tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
				return
			}
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
				return
			}
			assert.NilError(t, err)

			if tt.expected != nil {
				assert.DeepEqual(
					t,
					tt.expected,
					result.CheckedExpr.GetExpr(),
					protocmp.Transform(),
					protocmp.IgnoreFields(&expr.Expr{}, "id"),
				)
			}
		})
	}
}

// TestFilter_WithMacros_ConcurrentSharedDeclarations verifies that calling
// WithMacros concurrently on filters that share the same *Declarations does
// not cause a data race. Run with -race to detect violations.
func TestFilter_WithMacros_ConcurrentSharedDeclarations(t *testing.T) {
	t.Parallel()
	decl, err := NewDeclarations(
		DeclareStandardFunctions(),
		DeclareIdent("name", TypeString),
	)
	assert.NilError(t, err)
	macro := func(cursor *Cursor) {
		identExpr := cursor.Expr().GetIdentExpr()
		if identExpr == nil || identExpr.GetName() != "name" {
			return
		}
		cursor.ReplaceWithDeclarations(
			Text("renamed"),
			[]DeclarationOption{DeclareIdent("renamed", TypeString)},
		)
	}
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			filter, ferr := ParseFilter(&mockRequest{filter: `name = "test"`}, decl)
			assert.NilError(t, ferr)
			_, merr := filter.WithMacros(macro)
			assert.NilError(t, merr)
		})
	}
	wg.Wait()
}

// TestFilter_WithMacros_Retry verifies that calling WithMacros repeatedly on
// the same source Filter is idempotent: each call produces an equivalent
// rewritten expression and the source's CheckedExpr is never mutated. This
// pattern occurs in practice when a Spanner read-write transaction is retried
// — the driver invokes the callback (and therefore WithMacros) again with the
// same source Filter, and the second attempt must observe a pristine tree.
func TestFilter_WithMacros_Retry(t *testing.T) {
	t.Parallel()
	decl, err := NewDeclarations(
		DeclareStandardFunctions(),
		DeclareIdent("name", TypeString),
	)
	assert.NilError(t, err)
	macro := func(cursor *Cursor) {
		identExpr := cursor.Expr().GetIdentExpr()
		if identExpr == nil || identExpr.GetName() != "name" {
			return
		}
		cursor.ReplaceWithDeclarations(
			Text("renamed"),
			[]DeclarationOption{DeclareIdent("renamed", TypeString)},
		)
	}
	source, err := ParseFilter(&mockRequest{filter: `name = "test"`}, decl)
	assert.NilError(t, err)
	// Snapshot the original expression so we can verify the source is never mutated.
	originalExpr := proto.CloneOf(source.CheckedExpr)
	want := Equals(Text("renamed"), String("test"))
	for i := range 3 {
		result, werr := source.WithMacros(macro)
		assert.NilError(t, werr, "WithMacros call %d", i)
		assert.DeepEqual(
			t,
			want,
			result.CheckedExpr.GetExpr(),
			protocmp.Transform(),
			protocmp.IgnoreFields(&expr.Expr{}, "id"),
		)
		// The source must remain pristine across calls.
		assert.DeepEqual(
			t,
			originalExpr,
			source.CheckedExpr,
			protocmp.Transform(),
		)
	}
}

// TestFilter_WithMacros_Chain verifies that WithMacros can be chained — feeding
// the result of one call back into another — and that each intermediate Filter
// retains its own expression tree.
func TestFilter_WithMacros_Chain(t *testing.T) {
	t.Parallel()
	decl, err := NewDeclarations(
		DeclareStandardFunctions(),
		DeclareIdent("name", TypeString),
	)
	assert.NilError(t, err)
	renameTo := func(from, to string) Macro {
		return func(cursor *Cursor) {
			identExpr := cursor.Expr().GetIdentExpr()
			if identExpr == nil || identExpr.GetName() != from {
				return
			}
			cursor.ReplaceWithDeclarations(
				Text(to),
				[]DeclarationOption{DeclareIdent(to, TypeString)},
			)
		}
	}
	source, err := ParseFilter(&mockRequest{filter: `name = "test"`}, decl)
	assert.NilError(t, err)
	step1, err := source.WithMacros(renameTo("name", "step1"))
	assert.NilError(t, err)
	step2, err := step1.WithMacros(renameTo("step1", "step2"))
	assert.NilError(t, err)

	assert.DeepEqual(
		t,
		Equals(Text("name"), String("test")),
		source.CheckedExpr.GetExpr(),
		protocmp.Transform(),
		protocmp.IgnoreFields(&expr.Expr{}, "id"),
	)
	assert.DeepEqual(
		t,
		Equals(Text("step1"), String("test")),
		step1.CheckedExpr.GetExpr(),
		protocmp.Transform(),
		protocmp.IgnoreFields(&expr.Expr{}, "id"),
	)
	assert.DeepEqual(
		t,
		Equals(Text("step2"), String("test")),
		step2.CheckedExpr.GetExpr(),
		protocmp.Transform(),
		protocmp.IgnoreFields(&expr.Expr{}, "id"),
	)
}
