package resourcename

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHasParent(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		test     string
		name     string
		parent   string
		expected bool
	}{
		{
			test:     "valid parent and child",
			name:     "shippers/1/sites/1",
			parent:   "shippers/1",
			expected: true,
		},

		{
			test:     "not parent of self",
			name:     "shippers/1/sites/1/settings",
			parent:   "shippers/1/sites/1/settings",
			expected: false,
		},

		{
			test:     "empty parent",
			name:     "shippers/1/sites/1",
			parent:   "",
			expected: false,
		},

		{
			test:     "empty child and empty parent",
			name:     "",
			parent:   "",
			expected: false,
		},

		{
			test:     "singleton child",
			name:     "shippers/1/sites/1/settings",
			parent:   "shippers/1/sites/1",
			expected: true,
		},

		{
			test:     "wildcard parent",
			name:     "shippers/1/sites/1",
			parent:   "shippers/-",
			expected: true,
		},

		{
			test:     "full child",
			name:     "//freight-example.einride.tech/shippers/1/sites/1",
			parent:   "shippers/-",
			expected: true,
		},

		{
			test:     "full parent",
			name:     "shippers/1/sites/1",
			parent:   "//freight-example.einride.tech/shippers/-",
			expected: true,
		},

		{
			test:     "full parent",
			name:     "shippers/1/sites/1",
			parent:   "//freight-example.einride.tech/shippers/-",
			expected: true,
		},

		{
			test:     "full parent and child with different service names",
			name:     "//other-example.einride.tech/shippers/1/sites/1",
			parent:   "//freight-example.einride.tech/shippers/-",
			expected: false,
		},
	} {
		tt := tt
		t.Run(tt.test, func(t *testing.T) {
			t.Parallel()
			assert.Assert(t, HasParent(tt.name, tt.parent) == tt.expected)
		})
	}
}
