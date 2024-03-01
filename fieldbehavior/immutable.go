package fieldbehavior

import (
	"fmt"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ValidateImmutableFieldsWithMask returns a validation error if the message
// or field mask contains a field that is immutable and a change to an immutable field is requested.
// This can be used when validating update requests and want to return
// INVALID_ARGUMENT to the user.
// If you want to ignore immutable fields rather than error then use ClearFields().
//
// See: https://aip.dev/203
func ValidateImmutableFieldsWithMask(m proto.Message, mask *fieldmaskpb.FieldMask) error {
	return validateImmutableFields(m.ProtoReflect(), mask, "")
}

func validateImmutableFields(m protoreflect.Message, mask *fieldmaskpb.FieldMask, path string) error {
	for i := 0; i < m.Descriptor().Fields().Len(); i++ {
		field := m.Descriptor().Fields().Get(i)
		currPath := path
		if len(currPath) > 0 {
			currPath += "."
		}

		currPath += string(field.Name())
		if isImmutable(field) && hasPath(mask, currPath) {
			return fmt.Errorf("field is immutable: %s", currPath)
		}

		if field.Kind() == protoreflect.MessageKind {
			value := m.Get(field)
			switch {
			case field.IsList():
				for i := 0; i < value.List().Len(); i++ {
					if err := validateImmutableFields(value.List().Get(i).Message(), mask, currPath); err != nil {
						return err
					}
				}
			case field.IsMap():
				if field.MapValue().Kind() != protoreflect.MessageKind {
					continue
				}
				var mapErr error
				value.Map().Range(func(_ protoreflect.MapKey, value protoreflect.Value) bool {
					if err := validateImmutableFields(value.Message(), mask, currPath); err != nil {
						mapErr = err
						return false
					}

					return true
				})
				if mapErr != nil {
					return mapErr
				}
			default:
				if err := validateImmutableFields(value.Message(), mask, currPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func isImmutable(field protoreflect.FieldDescriptor) bool {
	return Has(field, annotations.FieldBehavior_IMMUTABLE)
}
