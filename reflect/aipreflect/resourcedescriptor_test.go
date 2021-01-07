package aipreflect

import (
	"testing"

	examplefreightv1 "go.einride.tech/aip/examples/proto/gen/einride/example/freight/v1"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"
)

func TestNewResourceDescriptor(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		msg           proto.Message
		expected      *ResourceDescriptor
		errorContains string
	}{
		{
			name: "einride.example.freight.v1.Shipper",
			msg:  &examplefreightv1.Shipper{},
			expected: &ResourceDescriptor{
				Type: "freight-example.einride.tech/Shipper",
				Names: []*ResourceNameDescriptor{
					{
						Type: "freight-example.einride.tech/Shipper",
						Pattern: ResourceNamePatternDescriptor{
							Segments: []ResourceNameSegmentDescriptor{
								{Value: "shippers"},
								{Value: "shipper", Variable: true},
							},
						},
					},
				},
			},
		},

		{
			name: "einride.example.freight.v1.Shipment",
			msg:  &examplefreightv1.Shipment{},
			expected: &ResourceDescriptor{
				Type: "freight-example.einride.tech/Shipment",
				Names: []*ResourceNameDescriptor{
					{
						Type: "freight-example.einride.tech/Shipment",
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
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			options := tt.msg.ProtoReflect().Descriptor().Options()
			protoDesc := proto.GetExtension(options, annotations.E_Resource).(*annotations.ResourceDescriptor)
			assert.Assert(t, protoDesc != nil)
			actual, err := NewResourceDescriptor(protoDesc)
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, tt.expected, actual)
			}
		})
	}
}
