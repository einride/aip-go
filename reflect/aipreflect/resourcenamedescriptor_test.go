package aipreflect

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestNewResourceNameDescriptor(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		pattern       string
		expected      *ResourceNameDescriptor
		errorContains string
	}{
		{
			pattern: "shippers/{shipper}",
			expected: &ResourceNameDescriptor{
				Pattern: ResourceNamePatternDescriptor{
					Segments: []ResourceNameSegmentDescriptor{
						{Value: "shippers"},
						{Value: "shipper", Variable: true},
					},
				},
			},
		},

		{
			pattern: "shippers/{shipper}/shipments/{shipment}",
			expected: &ResourceNameDescriptor{
				Pattern: ResourceNamePatternDescriptor{
					Segments: []ResourceNameSegmentDescriptor{
						{Value: "shippers"},
						{Value: "shipper", Variable: true},
						{Value: "shipments"},
						{Value: "shipment", Variable: true},
					},
				},
			},
		},
	} {
		tt := tt
		t.Run(tt.pattern, func(t *testing.T) {
			t.Parallel()
			actual, err := NewResourceNameDescriptor(tt.pattern)
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, tt.expected, actual)
			}
		})
	}
}
