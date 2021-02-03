package fieldmask

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Update updates fields in dst with values from src according to the provided field mask.
// Nested messages are recursively updated in the same manner.
// Repeated fields and maps are copied by reference from src to dst.
//
// If no update mask is provided, only non-zero values of src are copied to dst.
// If the special value "*" is provided as the field mask, a full replacement of all fields in dst is done.
//
// See: https://google.aip.dev/134 (Standard methods: Update).
func Update(mask *fieldmaskpb.FieldMask, dst, src proto.Message) {
	dstReflect := dst.ProtoReflect()
	srcReflect := src.ProtoReflect()
	if dstReflect.Descriptor() != srcReflect.Descriptor() {
		panic(fmt.Sprintf(
			"dst (%s) and src (%s) messages have different types",
			dstReflect.Descriptor().FullName(),
			srcReflect.Descriptor().FullName(),
		))
	}
	switch {
	// Special-case: No update mask.
	// Update all fields of src that are set on the wire.
	case len(mask.GetPaths()) == 0:
		updateWireSetFields(dstReflect, srcReflect)
	// Special-case: Update mask is [*].
	// Do a full replacement of all fields.
	case len(mask.GetPaths()) == 1 && mask.GetPaths()[0] == "*":
		proto.Reset(dst)
		proto.Merge(dst, src)
	default:
		for _, path := range mask.GetPaths() {
			segments := strings.Split(path, ".")
			updateNamedField(dstReflect, srcReflect, segments)
		}
	}
}

func updateWireSetFields(dst, src protoreflect.Message) {
	// handle oneof fields that have changed between src and dst
	oneofs := dst.Descriptor().Oneofs()
	for i := 0; i < oneofs.Len(); i++ {
		oneof := oneofs.Get(i)
		whichDst := dst.WhichOneof(oneof)
		whichSrc := src.WhichOneof(oneof)
		if whichDst != whichSrc && whichSrc != nil {
			dst.Set(whichSrc, src.Get(whichSrc))
		}
	}
	src.Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		if isMessage(field) {
			updateWireSetFields(dst.Get(field).Message(), value.Message())
			return true
		}
		dst.Set(field, value)
		return true
	})
}

func updateNamedField(dst, src protoreflect.Message, segments []string) {
	if len(segments) == 0 {
		return
	}
	field := src.Descriptor().Fields().ByName(protoreflect.Name(segments[0]))
	if field == nil {
		return // no known field by that name
	}
	// a field in this message
	if len(segments) == 1 {
		dst.Set(field, src.Get(field))
		return
	}
	if !isMessage(field) {
		// not a message so can not have a field with that name
		return
	}
	updateNamedField(dst.Get(field).Message(), src.Get(field).Message(), segments[1:])
}

func isMessage(field protoreflect.FieldDescriptor) bool {
	return (field.Kind() == protoreflect.MessageKind || field.Kind() == protoreflect.GroupKind) &&
		!field.IsMap() && !field.IsList()
}
