package fieldbehavior

import (
	"fmt"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ValidateRequiredFields returns a validation error if any field annotated as required does not have a value.
// See: https://aip.dev/203
func ValidateRequiredFields(m proto.Message) error {
	return validateRequiredFields(m.ProtoReflect(), nil, "")
}

func ValidateRequiredFieldsWithMask(m proto.Message, mask *fieldmaskpb.FieldMask) error {
	return validateRequiredFields(m.ProtoReflect(), mask, "")
}

func validateRequiredFields(reflectMessage protoreflect.Message, mask *fieldmaskpb.FieldMask, path string) error {
	for i := 0; i < reflectMessage.Descriptor().Fields().Len(); i++ {
		field := reflectMessage.Descriptor().Fields().Get(i)
		currPath := path
		if len(currPath) > 0 {
			currPath += "."
		}
		currPath += string(field.Name())
		if !isPresent(reflectMessage, field) {
			if Has(field, annotations.FieldBehavior_REQUIRED) &&
				(len(mask.GetPaths()) == 0 || hasPath(mask, currPath)) {
				return fmt.Errorf("missing required field: %s", currPath)
			}
		} else if field.Kind() == protoreflect.MessageKind {
			value := reflectMessage.Get(field)
			if field.Cardinality() == protoreflect.Repeated {
				for i := 0; i < value.List().Len(); i++ {
					if err := validateRequiredFields(value.List().Get(i).Message(), mask, currPath); err != nil {
						return err
					}
				}
			} else if err := validateRequiredFields(value.Message(), mask, currPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func hasPath(mask *fieldmaskpb.FieldMask, needle string) bool {
	for _, straw := range mask.GetPaths() {
		if straw == needle {
			return true
		}
	}
	return false
}
