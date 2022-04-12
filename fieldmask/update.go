package fieldmask

import (
	"bytes"
	"fmt"
	"strings"

	"go.einride.tech/aip/fieldbehavior"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Update updates fields in dst with values from src according to the provided field mask.
// Nested messages are recursively updated in the same manner.
// Repeated fields and maps are copied by reference from src to dst.
// Field mask paths referring to Individual entries in maps or
// repeated fields are ignored.
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
	case IsFullReplacement(mask):
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
	src.Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		switch {
		case field.IsList():
			dst.Set(field, value)
		case field.IsMap():
			dst.Set(field, value)
		case field.Message() != nil && !dst.Has(field):
			dst.Set(field, value)
		case field.Message() != nil:
			updateWireSetFields(dst.Get(field).Message(), value.Message())
		default:
			dst.Set(field, value)
		}
		return true
	})
}

func updateNamedField(dst, src protoreflect.Message, segments []string) {
	if len(segments) == 0 {
		return
	}
	field := src.Descriptor().Fields().ByName(protoreflect.Name(segments[0]))
	if field == nil {
		// no known field by that name
		return
	}
	if fieldbehavior.Has(field, annotations.FieldBehavior_OUTPUT_ONLY) {
		return
	}
	// a named field in this message
	if len(segments) == 1 {
		if fieldbehavior.Has(field, annotations.FieldBehavior_IMMUTABLE) {
			srcFld := src.Get(field)
			dstFld := dst.Get(field)
			cmp := true
			switch field.Kind() {
			case protoreflect.EnumKind:
				cmp = srcFld.Enum() != dstFld.Enum()
			case protoreflect.BoolKind:
				cmp = srcFld.Bool() != dstFld.Bool()
			case protoreflect.Int32Kind,
				protoreflect.Sint32Kind,
				protoreflect.Int64Kind,
				protoreflect.Sint64Kind,
				protoreflect.Sfixed32Kind,
				protoreflect.Sfixed64Kind:
				cmp = srcFld.Int() == dstFld.Int()
			case protoreflect.Uint32Kind,
				protoreflect.Uint64Kind,
				protoreflect.Fixed32Kind,
				protoreflect.Fixed64Kind:
				cmp = srcFld.Uint() == dstFld.Uint()
			case protoreflect.FloatKind,
				protoreflect.DoubleKind:
				cmp = srcFld.Float() == dstFld.Float()
			case protoreflect.StringKind:
				cmp = srcFld.String() == dstFld.String()
			case protoreflect.BytesKind:
				cmp = bytes.Equal(srcFld.Bytes(), dstFld.Bytes())
			}
			if !cmp {
				return
			}
		}
		if !src.Has(field) {
			dst.Clear(field)
		} else {
			dst.Set(field, src.Get(field))
		}
		return
	}

	// a named field in a nested message
	switch {
	case field.IsList(), field.IsMap():
		// nested fields in repeated or map not supported
		return
	case field.Message() != nil:
		// if message field is not set, allocate an empty value
		if !dst.Has(field) {
			dst.Set(field, dst.NewField(field))
		}
		if !src.Has(field) {
			src.Set(field, src.NewField(field))
		}
		updateNamedField(dst.Get(field).Message(), src.Get(field).Message(), segments[1:])
	default:
		return
	}
}
