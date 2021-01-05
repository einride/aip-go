package aipreflect

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestResourceNamePatternDescriptor_MarshalResourceName(t *testing.T) {
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
			pattern, err := NewResourceNamePatternDescriptor(tt.pattern)
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

func TestResourceNamePatternDescriptor_ValidateResourceName(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		pattern       string
		input         string
		errorContains string
	}{
		{
			name:    "ok, no variables",
			pattern: "singleton",
			input:   "singleton",
		},
		{
			name:    "ok, single-variable",
			pattern: "publishers/{publisher}",
			input:   "publishers/1",
		},
		{
			name:    "ok, multi-variable",
			pattern: "publishers/{publisher}/books/{book}",
			input:   "publishers/1/books/2",
		},
		{
			name:          "error, no variables",
			pattern:       "singleton",
			input:         "foo",
			errorContains: "expected segment 1 to be `singleton` but got `foo`",
		},
		{
			name:          "error, single variable, too short",
			pattern:       "publishers/{publisher}",
			input:         "publishers",
			errorContains: "expected 2 segments but got 1",
		},
		{
			name:          "error, single variable, too long",
			pattern:       "publishers/{publisher}",
			input:         "publishers/1/books",
			errorContains: "expected 2 segments but got 3",
		},
		{
			name:          "error, single variable, empty",
			pattern:       "publishers/{publisher}",
			input:         "publishers/",
			errorContains: "segment {publisher} is empty",
		},
		{
			name:          "error, multiple variables, too short",
			pattern:       "publishers/{publisher}/books/{book}",
			input:         "publishers/1",
			errorContains: "expected 4 segments but got 2",
		},
		{
			name:          "error, multiple variables, too long",
			pattern:       "publishers/{publisher}/books/{book}",
			input:         "publishers/1/books/2/foo",
			errorContains: "expected 4 segments but got 5",
		},
		{
			name:          "error, multiple variables, empty",
			pattern:       "publishers/{publisher}/books/{book}",
			input:         "publishers//books/1",
			errorContains: "segment {publisher} is empty",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pattern, err := NewResourceNamePatternDescriptor(tt.pattern)
			assert.NilError(t, err)
			if err := pattern.ValidateResourceName(tt.input); tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
