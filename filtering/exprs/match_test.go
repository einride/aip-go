package exprs

import (
	"testing"

	"go.einride.tech/aip/filtering"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
)

func TestMatch(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		expr     *expr.Expr
		matcher  Matcher
		expected bool
	}{
		{
			name:     "string: match",
			matcher:  MatchString("string"),
			expr:     filtering.String("string"),
			expected: true,
		},
		{
			name:    "string: another expr",
			matcher: MatchString("string"),
			expr:    filtering.Text("state"),
		},
		{
			name:    "string: wrong string",
			matcher: MatchString("string"),
			expr:    filtering.String("another string"),
		},
		{
			name:     "float: match",
			matcher:  MatchFloat(3.14),
			expr:     filtering.Float(3.14),
			expected: true,
		},
		{
			name:    "float: another expr",
			matcher: MatchFloat(3.14),
			expr:    filtering.Text("state"),
		},
		{
			name:    "float: wrong float",
			matcher: MatchFloat(3.14),
			expr:    filtering.Float(1.23),
		},
		{
			name:     "int: match",
			matcher:  MatchInt(3),
			expr:     filtering.Int(3),
			expected: true,
		},
		{
			name:    "int: another expr",
			matcher: MatchInt(3),
			expr:    filtering.Text("state"),
		},
		{
			name:    "int: wrong int",
			matcher: MatchInt(3),
			expr:    filtering.Int(1),
		},
		{
			name:     "text: match",
			matcher:  MatchText("text"),
			expr:     filtering.Text("text"),
			expected: true,
		},
		{
			name:    "text: another expr",
			matcher: MatchText("text"),
			expr:    filtering.Text("state"),
		},
		{
			name:    "text: wrong text",
			matcher: MatchText("text"),
			expr:    filtering.Text("another_text"),
		},
		{
			name:     "member: match",
			matcher:  MatchMember(MatchText("operand"), "field"),
			expr:     filtering.Member(filtering.Text("operand"), "field"),
			expected: true,
		},
		{
			name:    "member: another expr",
			matcher: MatchMember(MatchText("operand"), "field"),
			expr:    filtering.Text("state"),
		},
		{
			name:    "member: wrong field",
			matcher: MatchMember(MatchText("operand"), "field"),
			expr:    filtering.Member(filtering.Text("operand"), "another_field"),
		},
		{
			name: "function: match",
			matcher: MatchFunction(
				"=",
				MatchString("lhs"),
				MatchString("rhs"),
			),
			expr: filtering.Function(
				"=",
				filtering.String("lhs"),
				filtering.String("rhs"),
			),
			expected: true,
		},
		{
			name: "function: another expr",
			matcher: MatchFunction(
				"=",
				MatchString("lhs"),
				MatchString("rhs"),
			),
			expr: filtering.Text("state"),
		},
		{
			name: "function: wrong number of args",
			matcher: MatchFunction(
				"=",
				MatchString("lhs"),
				MatchString("rhs"),
			),
			expr: filtering.Function(
				"=",
				filtering.String("lhs"),
			),
		},
		{
			name: "function: wrong name",
			matcher: MatchFunction(
				"=",
				MatchString("lhs"),
				MatchString("rhs"),
			),
			expr: filtering.Function(
				"*",
				filtering.String("lhs"),
				filtering.String("rhs"),
			),
		},
		{
			name: "function: wrong number of args",
			matcher: MatchFunction(
				"=",
				MatchString("lhs"),
				MatchString("rhs"),
			),
			expr: filtering.Function(
				"=",
				filtering.String("lhs"),
			),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.matcher(tt.expr)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMatchAny(t *testing.T) {
	t.Parallel()
	t.Run("String", func(t *testing.T) {
		t.Parallel()
		var val string
		matcher := MatchAnyString(&val)
		exp := filtering.String("x")

		assert.Check(t, matcher(exp))
		assert.Equal(t, "x", val)
	})
	t.Run("Float", func(t *testing.T) {
		t.Parallel()
		var val float64
		matcher := MatchAnyFloat(&val)
		exp := filtering.Float(3.14)

		assert.Check(t, matcher(exp))
		assert.Equal(t, 3.14, val)
	})
	t.Run("Int", func(t *testing.T) {
		t.Parallel()
		var val int64
		matcher := MatchAnyInt(&val)
		exp := filtering.Int(3)

		assert.Check(t, matcher(exp))
		assert.Equal(t, int64(3), val)
	})
	t.Run("Text", func(t *testing.T) {
		t.Parallel()
		var val string
		matcher := MatchAnyText(&val)
		exp := filtering.Text("x")

		assert.Check(t, matcher(exp))
		assert.Equal(t, "x", val)
	})
	t.Run("Member", func(t *testing.T) {
		t.Parallel()
		var operand, field string
		matcher := MatchAnyMember(MatchAnyText(&operand), &field)
		exp := filtering.Member(filtering.Text("operand"), "field")

		assert.Check(t, matcher(exp))
		assert.Equal(t, "operand", operand)
		assert.Equal(t, "field", field)
	})
	t.Run("Function", func(t *testing.T) {
		t.Parallel()
		var fn, lhs, rhs string
		matcher := MatchAnyFunction(&fn, MatchAnyString(&lhs), MatchAnyString(&rhs))
		exp := filtering.Function("=", filtering.String("lhs"), filtering.String("rhs"))

		assert.Check(t, matcher(exp))
		assert.Equal(t, "=", fn)
		assert.Equal(t, "lhs", lhs)
		assert.Equal(t, "rhs", rhs)
	})
	t.Run("Any", func(t *testing.T) {
		t.Parallel()
		ex := &expr.Expr{}
		matcher := MatchAny(&ex)
		exp := filtering.Function("=", filtering.String("lhs"), filtering.String("rhs"))

		assert.Check(t, matcher(exp))
		assert.DeepEqual(t, exp, ex, protocmp.Transform())
	})
}
