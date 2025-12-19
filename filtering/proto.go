package filtering

import (
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type FilterOption func(opts *filterOptions)

type filterOptions struct {
	filterableFields []string
}

// WithFilterableFields marks the given fields as filterable.
// To mark a simple field (native types) or a message field (with all its underlying fields, recursively)
// as filterable, use the field name.
// To mark a specific nested field as filterable, use the full path of the field using the dot notation.
// For example:
//   - WithFilterableFields("string_field") marks the string_field as filterable.
//   - WithFilterableFields("nested_message") marks the nested_message field and all its underlying fields as
//     filterable.
//   - WithFilterableFields("nested_message.nested_string") marks the nested_string field as filterable.
//   - WithFilterableFields("nested_message.nested_string", "nested_message.nested_int32") marks the nested_string and
//     nested_int32 fields as filterable.
//
// EXPERIMENTAL: This option is experimental and may be changed or removed in the future.
func WithFilterableFields(fields ...string) FilterOption {
	return func(opts *filterOptions) {
		opts.filterableFields = fields
	}
}

// ProtoDeclarations returns declarations for all fields marked as filterable in the proto message.
// By default, no fields are marked as filterable. To mark a field as filterable, use the WithFilterableFields option.
// EXPERIMENTAL: This function is experimental and may be changed or removed in the future.
func ProtoDeclarations(msg proto.Message, opts ...FilterOption) (*Declarations, error) {
	options := filterOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	declOpts := []DeclarationOption{
		DeclareStandardFunctions(),
	}
	declOpts = append(declOpts, messageOptions(msg.ProtoReflect().Descriptor(), "", options)...)
	return NewDeclarations(
		declOpts...,
	)
}

func messageOptions(
	msg protoreflect.MessageDescriptor,
	path string,
	options filterOptions,
) []DeclarationOption {
	var opts []DeclarationOption
	for i := 0; i < msg.Fields().Len(); i++ {
		field := msg.Fields().Get(i)
		currPath := path
		if len(currPath) > 0 {
			currPath += "."
		}
		currPath += string(field.Name())

		// If filterable fields were explicitly set, check if this field is allowed
		var found bool
		for _, filter := range options.filterableFields {
			// Check if current path matches the filter exactly
			if currPath == filter {
				found = true
				break
			}
			// Check if current path is a prefix of the filter (for nested fields)
			// For example: filter="nested_message.nested_string" should match path="nested_message"
			if strings.HasPrefix(filter, currPath+".") {
				found = true
				break
			}
			// Check if filter is a prefix of current path (for allowing access to nested fields)
			// For example: filter="nested_message" should match path="nested_message.nested_string"
			if strings.HasPrefix(currPath, filter+".") {
				found = true
				break
			}
		}
		if !found {
			// Skip non-filterable fields
			continue
		}

		if field.IsList() {
			// TODO: Add support for lists?
			continue
		}
		if field.IsMap() {
			// TODO: Add support for maps?
			continue
		}

		switch field.Kind() {
		case protoreflect.StringKind:
			opts = append(opts, DeclareIdent(currPath, TypeString))
		case protoreflect.EnumKind:
			// Use proper enum type declaration for better type safety and validation
			enumType := dynamicpb.NewEnumType(field.Enum())
			opts = append(opts, DeclareEnumIdent(currPath, enumType))
		case protoreflect.BoolKind:
			opts = append(opts, DeclareIdent(currPath, TypeBool))
		case protoreflect.Int32Kind,
			protoreflect.Sint32Kind,
			protoreflect.Int64Kind,
			protoreflect.Sint64Kind,
			protoreflect.Sfixed32Kind,
			protoreflect.Sfixed64Kind:
			opts = append(opts, DeclareIdent(currPath, TypeInt))
		case protoreflect.Uint32Kind,
			protoreflect.Uint64Kind,
			protoreflect.Fixed32Kind,
			protoreflect.Fixed64Kind:
			// TODO: Can we support uint?
			opts = append(opts, DeclareIdent(currPath, TypeInt))
		case protoreflect.FloatKind,
			protoreflect.DoubleKind:
			opts = append(opts, DeclareIdent(currPath, TypeFloat))
		case protoreflect.BytesKind:
			// TODO: Can we support bytes?
			opts = append(opts, DeclareIdent(currPath, TypeString))
		case protoreflect.MessageKind:
			// Special handling for well-known types
			if field.Message().FullName() == "google.protobuf.Timestamp" {
				opts = append(opts, DeclareIdent(currPath, TypeTimestamp))
			} else {
				// For nested messages, recursively process their fields
				// but pass the same filterable field options so nested fields are filtered correctly
				fieldOpts := messageOptions(field.Message(), currPath, options)
				opts = append(opts, fieldOpts...)
			}
		case protoreflect.GroupKind:
			// TODO: Add support for groups?
			continue
		default:
			continue
		}
	}
	return opts
}
