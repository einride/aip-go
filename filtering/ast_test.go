package filtering

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestExpression_Filter(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		expression Expression
		expected   string
	}{
		{
			expression: Expression{
				Sequences: []Sequence{
					{
						Factors: []Factor{
							{Terms: []Term{{Simple: Restriction{Comparable: Member{Value: Text("a")}}}}},
							{Terms: []Term{{Simple: Restriction{Comparable: Member{Value: Text("b")}}}}},
						},
					},
					{Factors: []Factor{{Terms: []Term{{Simple: Restriction{Comparable: Member{Value: Text("c")}}}}}}},
					{Factors: []Factor{{Terms: []Term{{Simple: Restriction{Comparable: Member{Value: Text("d")}}}}}}},
				},
			},
			expected: "a b AND c AND d",
		},
	} {
		tt := tt
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.expression.Filter())
		})
	}
}

func TestTerm_Filter(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		term     Term
		expected string
	}{
		{
			term:     Term{Minus: true, Simple: Restriction{Comparable: Member{Value: Text("30")}}},
			expected: "-30",
		},
	} {
		tt := tt
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.term.Filter())
		})
	}
}
