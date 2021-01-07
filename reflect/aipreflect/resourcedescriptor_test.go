package aipreflect

import (
	"testing"

	examplefreightv1 "go.einride.tech/aip/examples/proto/gen/einride/example/freight/v1"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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
				Singular: "shipper",
				Plural:   "shippers",
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
				Singular: "shipment",
				Plural:   "shipments",
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

func TestResourceDescriptor_InferMethodName(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		resource      *ResourceDescriptor
		methodType    MethodType
		expected      protoreflect.Name
		errorContains string
	}{
		{
			name: "get",
			resource: &ResourceDescriptor{
				Singular: "userEvent",
				Plural:   "userEvents",
			},
			methodType: MethodTypeGet,
			expected:   "GetUserEvent",
		},

		{
			name: "list",
			resource: &ResourceDescriptor{
				Singular: "yellowSubmarine",
				Plural:   "yellowSubmarines",
			},
			methodType: MethodTypeList,
			expected:   "ListYellowSubmarines",
		},

		{
			name: "missing singular",
			resource: &ResourceDescriptor{
				Plural: "userEvents",
			},
			methodType:    MethodTypeGet,
			errorContains: "singular not specified",
		},

		{
			name: "missing plural",
			resource: &ResourceDescriptor{
				Singular: "userEvent",
			},
			methodType:    MethodTypeList,
			errorContains: "plural not specified",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual, err := tt.resource.InferMethodName(tt.methodType)
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
				assert.Equal(t, protoreflect.Name(""), actual)
			} else {
				assert.NilError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}
