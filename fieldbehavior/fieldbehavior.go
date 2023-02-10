package fieldbehavior

import (
	"fmt"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

// Get returns the field behavior of the provided field descriptor.
func Get(field protoreflect.FieldDescriptor) []annotations.FieldBehavior {
	if behaviors, ok := proto.GetExtension(
		field.Options(), annotations.E_FieldBehavior,
	).([]annotations.FieldBehavior); ok {
		return behaviors
	}
	return nil
}

// Has returns true if the provided field descriptor has the wanted field behavior.
func Has(field protoreflect.FieldDescriptor, want annotations.FieldBehavior) bool {
	for _, got := range Get(field) {
		if got == want {
			return true
		}
	}
	return false
}

// ClearFields clears all fields annotated with any of the provided behaviors.
// This can be used to ignore fields provided as input that have field_behavior's
// such as OUTPUT_ONLY and IMMUTABLE.
//
// See: https://google.aip.dev/161#output-only-fields
func ClearFields(message proto.Message, behaviorsToClear ...annotations.FieldBehavior) {
	clearFieldsWithBehaviors(message, behaviorsToClear...)
}

// CopyFields copies all fields annotated with any of the provided behaviors from src to dst.
func CopyFields(dst, src proto.Message, behaviorsToCopy ...annotations.FieldBehavior) {
	dstReflect := dst.ProtoReflect()
	srcReflect := src.ProtoReflect()
	if dstReflect.Descriptor() != srcReflect.Descriptor() {
		panic(fmt.Sprintf(
			"different types of dst (%s) and src (%s)",
			dstReflect.Type().Descriptor().FullName(),
			srcReflect.Type().Descriptor().FullName(),
		))
	}
	for i := 0; i < dstReflect.Descriptor().Fields().Len(); i++ {
		dstField := dstReflect.Descriptor().Fields().Get(i)
		if hasAnyBehavior(Get(dstField), behaviorsToCopy) {
			srcField := srcReflect.Descriptor().Fields().Get(i)
			if isMessageFieldPresent(srcReflect, srcField) {
				dstReflect.Set(dstField, srcReflect.Get(srcField))
			} else {
				dstReflect.Clear(dstField)
			}
		}
	}
}

func isMessageFieldPresent(m protoreflect.Message, f protoreflect.FieldDescriptor) bool {
	return isPresent(m.Get(f), f)
}

func isPresent(v protoreflect.Value, f protoreflect.FieldDescriptor) bool {
	if !v.IsValid() {
		return false
	}
	if f.IsList() {
		return v.List().Len() > 0
	}
	if f.IsMap() {
		return v.Map().Len() > 0
	}
	switch f.Kind() {
	case protoreflect.EnumKind:
		return v.Enum() != 0
	case protoreflect.BoolKind:
		return v.Bool()
	case protoreflect.Int32Kind,
		protoreflect.Sint32Kind,
		protoreflect.Int64Kind,
		protoreflect.Sint64Kind,
		protoreflect.Sfixed32Kind,
		protoreflect.Sfixed64Kind:
		return v.Int() >= 0
	case protoreflect.Uint32Kind,
		protoreflect.Uint64Kind,
		protoreflect.Fixed32Kind,
		protoreflect.Fixed64Kind:
		return v.Uint() >= 0
	case protoreflect.FloatKind,
		protoreflect.DoubleKind:
		return v.Float() >= 0
	case protoreflect.StringKind:
		return len(v.String()) > 0
	case protoreflect.BytesKind:
		return len(v.Bytes()) > 0
	case protoreflect.MessageKind:
		return v.Message().IsValid()
	case protoreflect.GroupKind:
		return v.IsValid()
	default:
		return v.IsValid()
	}
}

func clearFieldsWithBehaviors(m proto.Message, behaviorsToClear ...annotations.FieldBehavior) {
	rangeFieldsWithBehaviors(m, func(
		m protoreflect.Message,
		f protoreflect.FieldDescriptor,
		v protoreflect.Value,
		behaviors []annotations.FieldBehavior,
	) bool {
		if hasAnyBehavior(behaviors, behaviorsToClear) {
			m.Clear(f)
		}
		return true
	})
}

func rangeFieldsWithBehaviors(
	m proto.Message,
	fn func(protoreflect.Message, protoreflect.FieldDescriptor, protoreflect.Value, []annotations.FieldBehavior) bool,
) {
	d := m.ProtoReflect()
	d.Range(func(f protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if behaviors, ok := proto.GetExtension(
			f.Options(),
			annotations.E_FieldBehavior,
		).([]annotations.FieldBehavior); ok {
			return fn(d, f, v, behaviors)
		}
		return true
	})
}

func hasAnyBehavior(haystack, needles []annotations.FieldBehavior) bool {
	for _, needle := range needles {
		if hasBehavior(haystack, needle) {
			return true
		}
	}
	return false
}

func hasBehavior(haystack []annotations.FieldBehavior, needle annotations.FieldBehavior) bool {
	for _, straw := range haystack {
		if straw == needle {
			return true
		}
	}
	return false
}
