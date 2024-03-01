package aipreflect

import (
	"fmt"
	"strings"

	"go.einride.tech/aip/resourcename"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protopath"
	"google.golang.org/protobuf/reflect/protorange"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// ValidateResourceReferences validates the resource reference fields in a message.
func ValidateResourceReferences(message proto.Message) error {
	return protorange.Range(message.ProtoReflect(), func(values protopath.Values) error {
		curr := values.Index(-1)
		var field protoreflect.FieldDescriptor
		var fieldValue string
		switch curr.Step.Kind() {
		case protopath.FieldAccessStep:
			field = curr.Step.FieldDescriptor()
			if field.Kind() != protoreflect.StringKind {
				return nil
			}
			if field.Cardinality() == protoreflect.Repeated {
				return nil
			}
			fieldValue = curr.Value.String()
		case protopath.ListIndexStep:
			prev := values.Index(-2)
			field = prev.Step.FieldDescriptor()
			if field.Kind() != protoreflect.StringKind {
				return nil
			}
			fieldValue = curr.Value.String()
		default:
			return nil
		}
		resourceReferenceAnnotation := proto.GetExtension(
			field.Options(), annotations.E_ResourceReference,
		).(*annotations.ResourceReference)
		if resourceReferenceAnnotation == nil {
			return nil
		}
		var errValidate error
		RangeResourceDescriptorsInPackage(
			protoregistry.GlobalFiles,
			field.ParentFile().Package(),
			func(resource *annotations.ResourceDescriptor) bool {
				if resource.GetType() != resourceReferenceAnnotation.GetType() {
					return true
				}
				for _, pattern := range resource.GetPattern() {
					if resourcename.Match(pattern, fieldValue) {
						return false
					}
				}
				errValidate = fmt.Errorf(
					"value '%s' of field %s is not a valid resource reference for %s",
					fieldValue,
					// trim the message type from the path
					strings.TrimLeft(strings.TrimLeftFunc(values.Path.String(), func(r rune) bool {
						return r != ')'
					}), ")."),
					resourceReferenceAnnotation.GetType(),
				)
				return false
			},
		)
		return errValidate
	})
}
