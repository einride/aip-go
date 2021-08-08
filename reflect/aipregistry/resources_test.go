package aipregistry

import (
	"testing"

	examplefreightv1 "go.einride.tech/aip/proto/gen/einride/example/freight/v1"
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
				examplefreightv1.File_einride_example_freight_v1_freight_service_proto,
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
					Methods: map[aipreflect.MethodType]protoreflect.FullName{
						aipreflect.MethodTypeGet:    "einride.example.freight.v1.FreightService.GetShipper",
						aipreflect.MethodTypeList:   "einride.example.freight.v1.FreightService.ListShippers",
						aipreflect.MethodTypeCreate: "einride.example.freight.v1.FreightService.CreateShipper",
						aipreflect.MethodTypeUpdate: "einride.example.freight.v1.FreightService.UpdateShipper",
						aipreflect.MethodTypeDelete: "einride.example.freight.v1.FreightService.DeleteShipper",
					},
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
					Methods: map[aipreflect.MethodType]protoreflect.FullName{
						aipreflect.MethodTypeGet:    "einride.example.freight.v1.FreightService.GetShipment",
						aipreflect.MethodTypeList:   "einride.example.freight.v1.FreightService.ListShipments",
						aipreflect.MethodTypeCreate: "einride.example.freight.v1.FreightService.CreateShipment",
						aipreflect.MethodTypeUpdate: "einride.example.freight.v1.FreightService.UpdateShipment",
						aipreflect.MethodTypeDelete: "einride.example.freight.v1.FreightService.DeleteShipment",
					},
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
					Methods: map[aipreflect.MethodType]protoreflect.FullName{
						aipreflect.MethodTypeGet:      "einride.example.freight.v1.FreightService.GetSite",
						aipreflect.MethodTypeList:     "einride.example.freight.v1.FreightService.ListSites",
						aipreflect.MethodTypeCreate:   "einride.example.freight.v1.FreightService.CreateSite",
						aipreflect.MethodTypeUpdate:   "einride.example.freight.v1.FreightService.UpdateSite",
						aipreflect.MethodTypeDelete:   "einride.example.freight.v1.FreightService.DeleteSite",
						aipreflect.MethodTypeBatchGet: "einride.example.freight.v1.FreightService.BatchGetSites",
					},
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
