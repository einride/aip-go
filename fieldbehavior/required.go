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
	return validateRequiredFields(
		m.ProtoReflect(),
		&fieldmaskpb.FieldMask{Paths: []string{"*"}},
		"",
	)
}

func ValidateRequiredFieldsWithMask(m proto.Message, mask *fieldmaskpb.FieldMask) error {
	return validateRequiredFields(m.ProtoReflect(), mask, "")
}

func validateRequiredFields(reflectMessage protoreflect.Message, mask *fieldmaskpb.FieldMask, path string) error {
	// If no paths are provided, the field mask should be treated to be equivalent
	// to all fields set on the wire. This means that no required fields can be missing,
	// since if they were missing they're not set on the wire.
	if len(mask.GetPaths()) == 0 {
		return nil
	}
	for i := 0; i < reflectMessage.Descriptor().Fields().Len(); i++ {
		field := reflectMessage.Descriptor().Fields().Get(i)
		currPath := path
		if len(currPath) > 0 {
			currPath += "."
		}
		currPath += string(field.Name())
		if !isMessageFieldPresent(reflectMessage, field) {
			if Has(field, annotations.FieldBehavior_REQUIRED) && hasPath(mask, currPath) {
				return fmt.Errorf("missing required field: %s", currPath)
			}
		} else if field.Kind() == protoreflect.MessageKind {
			value := reflectMessage.Get(field)
			switch {
			case field.IsList():
				for i := 0; i < value.List().Len(); i++ {
					if err := validateRequiredFields(value.List().Get(i).Message(), mask, currPath); err != nil {
						return err
					}
				}
			case field.IsMap():
				if field.MapValue().Kind() != protoreflect.MessageKind {
					continue
				}
				var mapErr error
				value.Map().Range(func(_ protoreflect.MapKey, value protoreflect.Value) bool {
					if err := validateRequiredFields(value.Message(), mask, currPath); err != nil {
						mapErr = err
						return false
					}

					return true
				})
				if mapErr != nil {
					return mapErr
				}
			default:
				if err := validateRequiredFields(value.Message(), mask, currPath); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func isEmpty(mask *fieldmaskpb.FieldMask) bool {
	return mask == nil || len(mask.GetPaths()) == 0
}

func hasPath(mask *fieldmaskpb.FieldMask, needle string) bool {
	if isEmpty(mask) {
		return true
	}
	for _, straw := range mask.GetPaths() {
		if straw == "*" || straw == needle {
			return true
		}
	}
	return false
}
