package filtering

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParser_ParseExpression(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		filter        string
		expected      Expression
		errorContains string
	}{
		{
			filter: "time.now()",
			expected: Expression{
				Sequences: []Sequence{
					{
						Factors: []Factor{
							{
								Terms: []Term{
									{
										Simple: Restriction{
											Comparable: Function{
												Names: []Name{Text("time"), Text("now")},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{filter: "<", errorContains: "expected value, got <"},
	} {
		tt := tt
		t.Run(tt.filter, func(t *testing.T) {
			t.Parallel()
			var parser Parser
			parser.Init(tt.filter)
			actual, err := parser.ParseExpression()
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, tt.expected, actual)
			}
		})
	}
}

func TestParser_ParseArg(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		filter        string
		expected      Arg
		errorContains string
	}{
		{filter: "time.now()", expected: Function{Names: []Name{Text("time"), Text("now")}}},
		{filter: "<", errorContains: "expected value, got <"},
	} {
		tt := tt
		t.Run(tt.filter, func(t *testing.T) {
			t.Parallel()
			var parser Parser
			parser.Init(tt.filter)
			actual, err := parser.ParseArg()
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, tt.expected, actual)
			}
		})
	}
}

func TestParser_ParseFunction(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		filter        string
		expected      Function
		errorContains string
	}{
		{filter: "time.now()", expected: Function{Names: []Name{Text("time"), Text("now")}}},
		{filter: "<", errorContains: "expected name, got <"},
	} {
		tt := tt
		t.Run(tt.filter, func(t *testing.T) {
			t.Parallel()
			var parser Parser
			parser.Init(tt.filter)
			actual, err := parser.ParseFunction()
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, tt.expected, actual)
			}
		})
	}
}

func TestParser_ParseComparable(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		filter        string
		expected      Comparable
		errorContains string
	}{
		{filter: "foo", expected: Member{Value: Text("foo")}},
		{filter: "<", errorContains: "expected value, got <"},
	} {
		tt := tt
		t.Run(tt.filter, func(t *testing.T) {
			t.Parallel()
			var parser Parser
			parser.Init(tt.filter)
			actual, err := parser.ParseComparable()
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, tt.expected, actual)
			}
		})
	}
}

func TestParser_ParseComparator(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		filter        string
		expected      Comparator
		errorContains string
	}{
		{filter: "<=", expected: ComparatorLessEquals},
		{filter: "<", expected: ComparatorLessThan},
		{filter: ">=", expected: ComparatorGreaterEquals},
		{filter: ">", expected: ComparatorGreaterThan},
		{filter: "!=", expected: ComparatorNotEquals},
		{filter: "=", expected: ComparatorEquals},
		{filter: ":", expected: ComparatorHas},
		{filter: "?", errorContains: "expected comparator, got TEXT"},
	} {
		tt := tt
		t.Run(tt.filter, func(t *testing.T) {
			t.Parallel()
			var parser Parser
			parser.Init(tt.filter)
			actual, err := parser.ParseComparator()
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, tt.expected, actual)
			}
		})
	}
}

func TestParser_ParseValue(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		filter        string
		expected      Value
		errorContains string
	}{
		{filter: "FOO", expected: Text("FOO")},
		{filter: "'bar'", expected: String("'bar'")},
		{filter: "-FOO", errorContains: "expected value, got -"},
	} {
		tt := tt
		t.Run(tt.filter, func(t *testing.T) {
			t.Parallel()
			var parser Parser
			parser.Init(tt.filter)
			actual, err := parser.ParseValue()
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, tt.expected, actual)
			}
		})
	}
}

func TestParser_ParseName(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		filter        string
		expected      Name
		errorContains string
	}{
		{filter: "NOT", expected: KeywordNot},
		{filter: "AND", expected: KeywordAnd},
		{filter: "OR", expected: KeywordOr},
		{filter: "FOO", expected: Text("FOO")},
		{filter: "-FOO", errorContains: "expected name, got -"},
	} {
		tt := tt
		t.Run(tt.filter, func(t *testing.T) {
			t.Parallel()
			var parser Parser
			parser.Init(tt.filter)
			actual, err := parser.ParseName()
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, tt.expected, actual)
			}
		})
	}
}

func TestParser_ParseKeyword(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		filter        string
		expected      Keyword
		errorContains string
	}{
		{filter: "NOT", expected: KeywordNot},
		{filter: "AND", expected: KeywordAnd},
		{filter: "OR", expected: KeywordOr},
		{filter: "FOO", errorContains: "expected keyword, got TEXT"},
	} {
		tt := tt
		t.Run(tt.filter, func(t *testing.T) {
			t.Parallel()
			var parser Parser
			parser.Init(tt.filter)
			actual, err := parser.ParseKeyword()
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, tt.expected, actual)
			}
		})
	}
}
