package resourcename

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestPattern_MarshalResourceName(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		pattern       string
		values        []string
		expected      string
		errorContains string
	}{
		{
			name:     "no variables",
			pattern:  "singleton",
			expected: "singleton",
		},
		{
			name:     "single variable",
			pattern:  "publishers/{publisher}",
			values:   []string{"1"},
			expected: "publishers/1",
		},
		{
			name:     "multiple variables",
			pattern:  "publishers/{publisher}/books/{book}",
			values:   []string{"1", "2"},
			expected: "publishers/1/books/2",
		},
		{
			name:          "too few values",
			pattern:       "publishers/{publisher}/books/{book}",
			values:        []string{"1"},
			errorContains: "got 1 values but expected 2",
		},
		{
			name:          "too many values",
			pattern:       "publishers/{publisher}/books/{book}",
			values:        []string{"1", "2", "3"},
			errorContains: "got 3 values but expected 2",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pattern, err := ParsePattern(tt.pattern)
			assert.NilError(t, err)
			actual, err := pattern.MarshalResourceName(tt.values...)
			assert.Equal(t, tt.expected, actual)
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

// nolint: gochecknoglobals
var sink string

func BenchmarkPattern_MarshalResourceName(b *testing.B) {
	pattern, err := ParsePattern("publishers/{publisher}/books/{book}")
	assert.NilError(b, err)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		name, _ := pattern.MarshalResourceName("1", "2")
		sink = name
	}
}
