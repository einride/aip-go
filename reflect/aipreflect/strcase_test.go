package aipreflect

import (
	"testing"

	"gotest.tools/v3/assert"
)

func Test_initialUpperCase(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		s        string
		expected string
	}{
		{s: "", expected: ""},
		{s: "a", expected: "A"},
		{s: "aaa", expected: "Aaa"},
	} {
		t.Run(tt.s, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, initialUpperCase(tt.s))
		})
	}
}
