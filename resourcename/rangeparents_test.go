package resourcename

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestRangeParents(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "empty",
			input: "",
		},

		{
			name:  "singleton",
			input: "foo",
		},

		{
			name:  "single",
			input: "foo/bar",
			expected: []string{
				"foo",
			},
		},

		{
			name:  "multiple",
			input: "foo/bar/baz/123",
			expected: []string{
				"foo",
				"foo/bar",
				"foo/bar/baz",
			},
		},

		{
			name:  "full",
			input: "//test.example.com/foo/bar/baz/123",
			expected: []string{
				"foo",
				"foo/bar",
				"foo/bar/baz",
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var actual []string
			RangeParents(tt.input, func(parent string) bool {
				actual = append(actual, parent)
				return true
			})
			assert.DeepEqual(t, tt.expected, actual)
		})
	}
}
