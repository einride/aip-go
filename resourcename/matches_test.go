package resourcename

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestMatches(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		test     string
		name     string
		pattern  string
		expected bool
	}{
		{
			test:     "valid pattern",
			name:     "shippers/1/sites/1",
			pattern:  "shippers/{shipper}/sites/{site}",
			expected: true,
		},

		{
			test:     "name longer than pattern",
			name:     "shippers/1/sites/1/settings",
			pattern:  "shippers/{shipper}/sites/{site}",
			expected: false,
		},

		{
			test:     "empty pattern",
			pattern:  "",
			name:     "shippers/1/sites/1",
			expected: false,
		},

		{
			test:     "empty pattern and empty name",
			pattern:  "",
			name:     "",
			expected: false,
		},

		{
			test:     "singleton",
			name:     "shippers/1/sites/1/settings",
			pattern:  "shippers/{shipper}/sites/{site}/settings",
			expected: true,
		},

		{
			test:     "wildcard pattern",
			name:     "shippers/1/sites/1",
			pattern:  "shippers/-/sites/-",
			expected: false,
		},

		{
			test:     "full parent",
			name:     "//freight-example.einride.tech/shippers/1/sites/1",
			pattern:  "shippers/{shipper}/sites/{site}",
			expected: true,
		},

		{
			test:     "full pattern",
			name:     "shippers/1",
			pattern:  "//freight-example.einride.tech/shippers/{shipper}",
			expected: false,
		},

		{
			test:     "slash prefix in the name",
			name:     "/shippers/1",
			pattern:  "shippers/{shipper}",
			expected: true,
		},

		{
			test:     "slash prefix in the pattern",
			name:     "shippers/1",
			pattern:  "/shippers/{shipper}",
			expected: true,
		},
	} {
		tt := tt
		t.Run(tt.test, func(t *testing.T) {
			t.Parallel()
			assert.Check(
				t,
				Match(tt.pattern, tt.name) == tt.expected,
				fmt.Sprintf("expected Match(%q, %q)=%t", tt.pattern, tt.name, tt.expected),
			)
		})
	}
}
