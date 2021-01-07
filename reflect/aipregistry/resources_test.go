package aipregistry

import (
	"testing"

	examplefreightv1 "go.einride.tech/aip/examples/proto/gen/einride/example/freight/v1"
	"go.einride.tech/aip/reflect/aipreflect"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"gotest.tools/v3/assert"
)

func TestNewResources(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		files    []protoreflect.FileDescriptor
		expected map[aipreflect.ResourceTypeName]*aipreflect.ResourceDescriptor
	}{
		{
			name: "einride.example.freight.v1",
			files: []protoreflect.FileDescriptor{
				examplefreightv1.File_einride_example_freight_v1_shipper_proto,
				examplefreightv1.File_einride_example_freight_v1_site_proto,
				examplefreightv1.File_einride_example_freight_v1_shipment_proto,
			},
			expected: map[aipreflect.ResourceTypeName]*aipreflect.ResourceDescriptor{
				"freight-example.einride.tech/Shipper": {
					ParentFile: "einride/example/freight/v1/shipper.proto",
					Message:    "einride.example.freight.v1.Shipper",
					Type:       "freight-example.einride.tech/Shipper",
					Names: []*aipreflect.ResourceNameDescriptor{
						{
							Type: "freight-example.einride.tech/Shipper",
							Pattern: aipreflect.ResourceNamePatternDescriptor{
								Segments: []aipreflect.ResourceNameSegmentDescriptor{
									{Value: "shippers"},
									{Value: "shipper", Variable: true},
								},
							},
						},
					},
					Singular: "shipper",
					Plural:   "shippers",
				},

				"freight-example.einride.tech/Shipment": {
					ParentFile: "einride/example/freight/v1/shipment.proto",
					Message:    "einride.example.freight.v1.Shipment",
					Type:       "freight-example.einride.tech/Shipment",
					Names: []*aipreflect.ResourceNameDescriptor{
						{
							Type: "freight-example.einride.tech/Shipment",
							Ancestors: []aipreflect.ResourceTypeName{
								"freight-example.einride.tech/Shipper",
							},
							Pattern: aipreflect.ResourceNamePatternDescriptor{
								Segments: []aipreflect.ResourceNameSegmentDescriptor{
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

				"freight-example.einride.tech/Site": {
					ParentFile: "einride/example/freight/v1/site.proto",
					Message:    "einride.example.freight.v1.Site",
					Type:       "freight-example.einride.tech/Site",
					Names: []*aipreflect.ResourceNameDescriptor{
						{
							Type: "freight-example.einride.tech/Site",
							Ancestors: []aipreflect.ResourceTypeName{
								"freight-example.einride.tech/Shipper",
							},
							Pattern: aipreflect.ResourceNamePatternDescriptor{
								Segments: []aipreflect.ResourceNameSegmentDescriptor{
									{Value: "shippers"},
									{Value: "shipper", Variable: true},
									{Value: "sites"},
									{Value: "site", Variable: true},
								},
							},
						},
					},
					Singular: "site",
					Plural:   "sites",
				},
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var files protoregistry.Files
			for _, file := range tt.files {
				assert.NilError(t, files.RegisterFile(file))
			}
			resources, err := NewResources(&files)
			assert.NilError(t, err)
			actual := map[aipreflect.ResourceTypeName]*aipreflect.ResourceDescriptor{}
			resources.RangeResources(func(resource *aipreflect.ResourceDescriptor) bool {
				actual[resource.Type] = resource
				return true
			})
			assert.DeepEqual(t, tt.expected, actual)
		})
	}
}
