package fieldmask

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// UpdateTopLevelFields updates the top-level fields in dst with values from src according to the provided field mask.
//
// Nested messages are copied by reference from src to dst.
// If no update mask is provided, only non-zero values of src are copied to dst.
// If the special value "*" is provided as the field mask, a full replacement of all fields in dst is done.
//
// See: https://google.aip.dev/134 (Standard methods: Update).
func UpdateTopLevelFields(mask *fieldmaskpb.FieldMask, dst, src proto.Message) {
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
		srcReflect.Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
			dstReflect.Set(field, value)
			return true
		})
	// Special-case: Update mask is [*].
	// Do a full replacement of all fields.
	case len(mask.GetPaths()) == 1 && mask.GetPaths()[0] == "*":
		proto.Reset(dst)
		proto.Merge(dst, src)
	default:
		fields := srcReflect.Descriptor().Fields()
		for _, path := range mask.GetPaths() {
			if isTopLevelField := !strings.ContainsRune(path, '.'); !isTopLevelField {
				continue // skip non-top-level fields
			}
			field := fields.ByName(protoreflect.Name(path))
			if field == nil {
				continue // no known field by that name
			}
			dstReflect.Set(field, srcReflect.Get(field))
		}
	}
}
