package fieldmask

import (
	"fmt"
	"reflect"
	"strings"

	"go.einride.tech/aip/fieldbehavior"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	updateModeWire int = 0
	updateModeFull int = 1
	updateModePath int = 2
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
func Update(mask *fieldmaskpb.FieldMask, dst, src proto.Message) error {
	dstReflect := dst.ProtoReflect()
	srcReflect := src.ProtoReflect()
	if dstReflect.Descriptor() != srcReflect.Descriptor() {
		panic(fmt.Sprintf(
			"dst (%s) and src (%s) messages have different types",
			dstReflect.Descriptor().FullName(),
			srcReflect.Descriptor().FullName(),
		))
	}
	var err error
	switch {
	case len(mask.GetPaths()) == 0:
		err = applyUpdate(dstReflect, srcReflect, updateModeWire)
	case IsFullReplacement(mask):
		err = applyUpdate(dstReflect, srcReflect, updateModeFull)
	default:
		for _, path := range mask.GetPaths() {
			segments := strings.Split(path, ".")
			err = applyUpdate(dstReflect, srcReflect, updateModePath, segments...)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func applyUpdate(dst, src protoreflect.Message, mode int, segments ...string) error {
	var err error
	if mode == updateModePath && len(segments) > 0 {
		field := src.Descriptor().Fields().ByName(protoreflect.Name(segments[0]))
		if field != nil && !fieldbehavior.Has(field, annotations.FieldBehavior_OUTPUT_ONLY) {
			if len(segments) == 1 {
				err = updateField(field, dst, src)
			} else {
				switch {
				case field.IsList(), field.IsMap():
					// nested fields in repeated or map not supported
				case field.Message() != nil:
					// if message field is not set, allocate an empty value
					if !dst.Has(field) {
						dst.Set(field, dst.NewField(field))
					}
					if !src.Has(field) {
						src.Set(field, src.NewField(field))
					}
					applyUpdate(dst.Get(field).Message(), src.Get(field).Message(), mode, segments[1:]...)
				}
			}
		}
	} else {
		src.Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
			if fieldbehavior.Has(field, annotations.FieldBehavior_OUTPUT_ONLY) {
				return true
			}
			if mode == updateModeFull {
				dst.Set(field, dst.NewField(field))
			}
			switch {
			case field.IsList(), field.IsMap():
				err = updateField(field, dst, src)
			case field.Message() != nil && !dst.Has(field):
				err = updateField(field, dst, src)
			case field.Message() != nil:
				err = applyUpdate(dst.Get(field).Message(), src.Get(field).Message(), mode)
			default:
				err = updateField(field, dst, src)
			}
			return err == nil
		})
	}
	return err
}

func updateField(field protoreflect.FieldDescriptor, dst, src protoreflect.Message) error {
	if fieldbehavior.Has(field, annotations.FieldBehavior_IMMUTABLE) {
		if !reflect.DeepEqual(src.Get(field).Interface(), dst.Get(field).Interface()) {
			return fmt.Errorf("immutable field altered")
		}
	}
	if !src.Has(field) {
		dst.Clear(field)
	} else {
		dst.Set(field, src.Get(field))
	}
	return nil
}
