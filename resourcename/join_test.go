package resourcename

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestJoin(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "zero",
			input:    []string{},
			expected: "/",
		},
		{
			name:     "one",
			input:    []string{"parent/1"},
			expected: "parent/1",
		},
		{
			name: "two",
			input: []string{
				"parent/1",
				"child/2",
			},
			expected: "parent/1/child/2",
		},
		{
			name: "root first",
			input: []string{
				"/",
				"child/2",
			},
			expected: "child/2",
		},
		{
			name: "root last",
			input: []string{
				"parent/1",
				"/",
			},
			expected: "parent/1",
		},
		{
			name: "root second",
			input: []string{
				"parent/1",
				"/",
				"child/2",
			},
			expected: "parent/1/child/2",
		},
		{
			name: "root first and last",
			input: []string{
				"/",
				"child/1",
				"/",
			},
			expected: "child/1",
		},
		{
			name: "only roots",
			input: []string{
				"/",
				"/",
			},
			expected: "/",
		},
		{
			name: "empty first",
			input: []string{
				"",
				"child/2",
			},
			expected: "child/2",
		},
		{
			name: "empty second",
			input: []string{
				"parent/1",
				"",
				"child/2",
			},
			expected: "parent/1/child/2",
		},
		{
			name: "invalid first suffix",
			input: []string{
				"parent/1/",
				"child/2",
			},
			expected: "parent/1/child/2",
		},
		{
			name: "invalid last suffix",
			input: []string{
				"parent/1",
				"child/2/",
			},
			expected: "parent/1/child/2",
		},

		{
			name: "fully qualified first",
			input: []string{
				"//foo.example.com/foo/1",
				"bar/2",
			},
			expected: "//foo.example.com/foo/1/bar/2",
		},
		{
			name: "fully qualified second",
			input: []string{
				"foo/1",
				"//foo.example.com/bar/2",
			},
			expected: "foo/1/bar/2",
		},
		{
			name: "fully qualified both",
			input: []string{
				"//foo.example.com/foo/1",
				"//bar.example.com/bar/2",
			},
			expected: "//foo.example.com/foo/1/bar/2",
		},

		// TODO: Should these be disallowed?
		// See https://github.com/einride/aip-go/pull/258
		{
			name: "first slash prefix",
			input: []string{
				"/parent/1",
				"child/2",
			},
			expected: "parent/1/child/2",
		},
		{
			name: "second slash prefix",
			input: []string{
				"parent/1",
				"/child/2",
			},
			expected: "parent/1/child/2",
		},
		{
			name: "thirds slash prefix",
			input: []string{
				"parent/1",
				"child/2",
				"/extra/3",
			},
			expected: "parent/1/child/2/extra/3",
		},
		{
			name: "all slash prefix",
			input: []string{
				"/parent/1",
				"/child/2",
				"/extra/3",
			},
			expected: "parent/1/child/2/extra/3",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual := Join(tt.input...)
			assert.Equal(t, actual, tt.expected)
		})
	}
}
