package resourcename

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestSprint(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		pattern   string
		variables []string
		expected  string
	}{
		{
			name:     "no variables",
			pattern:  "singleton",
			expected: "singleton",
		},

		{
			name:      "too many variables",
			pattern:   "singleton",
			variables: []string{"foo"},
			expected:  "singleton",
		},

		{
			name:      "single variable",
			pattern:   "publishers/{publisher}",
			variables: []string{"foo"},
			expected:  "publishers/foo",
		},

		{
			name:      "two variables",
			pattern:   "publishers/{publisher}/books/{book}",
			variables: []string{"foo", "bar"},
			expected:  "publishers/foo/books/bar",
		},

		{
			name:      "singleton two variables",
			pattern:   "publishers/{publisher}/books/{book}/settings",
			variables: []string{"foo", "bar"},
			expected:  "publishers/foo/books/bar/settings",
		},

		{
			name:      "empty variable",
			pattern:   "publishers/{publisher}/books/{book}",
			variables: []string{"foo", ""},
			expected:  "publishers/foo/books/",
		},

		{
			name:      "too few variables",
			pattern:   "publishers/{publisher}/books/{book}/settings",
			variables: []string{"foo"},
			expected:  "publishers/foo/books//settings",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, Sprint(tt.pattern, tt.variables...))
		})
	}
}

//nolint:gochecknoglobals
var benchmarkSprintSink string

func BenchmarkSprint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkSprintSink = Sprint("publishers/{publisher}/books/{book}", "foo", "bar")
	}
}
