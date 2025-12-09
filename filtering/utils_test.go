package filtering

import (
	"testing"

	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"gotest.tools/v3/assert"
)

// fullProtobufMessage creates a comprehensive dynamic protobuf message for testing which includes
// all supported protocol buffers field types.
// EXPERIMENTAL: This function is experimental and may be changed or removed in the future.
func fullProtobufMessage(t *testing.T) *dynamicpb.Message {
	// Create enum descriptor
	enumDesc := &descriptorpb.EnumDescriptorProto{
		Name: toPtr("TestEnum"),
		Value: []*descriptorpb.EnumValueDescriptorProto{
			{Name: toPtr("ENUM_VALUE_ZERO"), Number: toPtr(int32(0))},
			{Name: toPtr("ENUM_VALUE_ONE"), Number: toPtr(int32(1))},
			{Name: toPtr("ENUM_VALUE_TWO"), Number: toPtr(int32(2))},
		},
	}

	// Create nested message descriptor
	deepNestedDesc := &descriptorpb.DescriptorProto{
		Name: toPtr("DeepNested"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   toPtr("deep_string"),
				Number: toPtr(int32(1)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_STRING),
			},
		},
	}

	nestedDesc := &descriptorpb.DescriptorProto{
		Name: toPtr("NestedMessage"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   toPtr("nested_string"),
				Number: toPtr(int32(1)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_STRING),
			},
			{
				Name:     toPtr("deep_nested"),
				Number:   toPtr(int32(2)),
				Type:     toPtr(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
				TypeName: toPtr(".test.TestMessage.NestedMessage.DeepNested"),
			},
		},
		NestedType: []*descriptorpb.DescriptorProto{deepNestedDesc},
	}

	// Create main message descriptor with all field types
	msgDesc := &descriptorpb.DescriptorProto{
		Name: toPtr("TestMessage"),
		Field: []*descriptorpb.FieldDescriptorProto{
			// String field
			{
				Name:   toPtr("string_field"),
				Number: toPtr(int32(1)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_STRING),
			},
			// Bool field
			{
				Name:   toPtr("bool_field"),
				Number: toPtr(int32(2)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_BOOL),
			},
			// Integer fields
			{
				Name:   toPtr("int32_field"),
				Number: toPtr(int32(3)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_INT32),
			},
			{
				Name:   toPtr("int64_field"),
				Number: toPtr(int32(4)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_INT64),
			},
			{
				Name:   toPtr("sint32_field"),
				Number: toPtr(int32(5)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_SINT32),
			},
			{
				Name:   toPtr("sint64_field"),
				Number: toPtr(int32(6)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_SINT64),
			},
			{
				Name:   toPtr("sfixed32_field"),
				Number: toPtr(int32(7)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_SFIXED32),
			},
			{
				Name:   toPtr("sfixed64_field"),
				Number: toPtr(int32(8)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_SFIXED64),
			},
			// Unsigned integer fields
			{
				Name:   toPtr("uint32_field"),
				Number: toPtr(int32(9)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_UINT32),
			},
			{
				Name:   toPtr("uint64_field"),
				Number: toPtr(int32(10)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_UINT64),
			},
			{
				Name:   toPtr("fixed32_field"),
				Number: toPtr(int32(11)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_FIXED32),
			},
			{
				Name:   toPtr("fixed64_field"),
				Number: toPtr(int32(12)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_FIXED64),
			},
			// Float fields
			{
				Name:   toPtr("float_field"),
				Number: toPtr(int32(13)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_FLOAT),
			},
			{
				Name:   toPtr("double_field"),
				Number: toPtr(int32(14)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_DOUBLE),
			},
			// Bytes field
			{
				Name:   toPtr("bytes_field"),
				Number: toPtr(int32(15)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_BYTES),
			},
			// Enum field
			{
				Name:     toPtr("enum_field"),
				Number:   toPtr(int32(16)),
				Type:     toPtr(descriptorpb.FieldDescriptorProto_TYPE_ENUM),
				TypeName: toPtr(".test.TestMessage.TestEnum"),
			},
			// Timestamp field (well-known type)
			{
				Name:     toPtr("timestamp_field"),
				Number:   toPtr(int32(17)),
				Type:     toPtr(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
				TypeName: toPtr(".google.protobuf.Timestamp"),
			},
			// Nested message field
			{
				Name:     toPtr("nested_message"),
				Number:   toPtr(int32(18)),
				Type:     toPtr(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
				TypeName: toPtr(".test.TestMessage.NestedMessage"),
			},
			// List fields (should be skipped)
			{
				Name:   toPtr("string_list"),
				Number: toPtr(int32(19)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_STRING),
				Label:  toPtr(descriptorpb.FieldDescriptorProto_LABEL_REPEATED),
			},
			// Map field (should be skipped) - maps are represented as repeated nested messages
			{
				Name:     toPtr("string_map"),
				Number:   toPtr(int32(20)),
				Type:     toPtr(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
				TypeName: toPtr(".test.TestMessage.StringMapEntry"),
				Label:    toPtr(descriptorpb.FieldDescriptorProto_LABEL_REPEATED),
			},
		},
		EnumType:   []*descriptorpb.EnumDescriptorProto{enumDesc},
		NestedType: []*descriptorpb.DescriptorProto{nestedDesc},
	}

	// Map entry message for the map field
	mapEntryDesc := &descriptorpb.DescriptorProto{
		Name: toPtr("StringMapEntry"),
		Options: &descriptorpb.MessageOptions{
			MapEntry: toPtr(true),
		},
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   toPtr("key"),
				Number: toPtr(int32(1)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_STRING),
			},
			{
				Name:   toPtr("value"),
				Number: toPtr(int32(2)),
				Type:   toPtr(descriptorpb.FieldDescriptorProto_TYPE_STRING),
			},
		},
	}
	msgDesc.NestedType = append(msgDesc.NestedType, mapEntryDesc)

	// Create file descriptor
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:        toPtr("test.proto"),
		Package:     toPtr("test"),
		MessageType: []*descriptorpb.DescriptorProto{msgDesc},
		Dependency:  []string{"google/protobuf/timestamp.proto"},
	}

	// Convert to protoreflect descriptor using global registry (includes well-known types)
	protoFile, err := protodesc.NewFile(fileDesc, protoregistry.GlobalFiles)
	assert.NilError(t, err)

	// Get the message descriptor
	messageDesc := protoFile.Messages().ByName("TestMessage")
	assert.Assert(t, messageDesc != nil)

	// Create dynamic message
	dynamicMsg := dynamicpb.NewMessage(messageDesc)
	return dynamicMsg
}

func toPtr[T any](v T) *T { return &v }
