package fieldmask

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Validate validates that the paths in the provided field mask are syntactically valid and
// refer to known fields in the specified message type.
func Validate(fm *fieldmaskpb.FieldMask, m proto.Message) error {
	// special case for '*'
	if stringsContain("*", fm.GetPaths()) {
		if len(fm.GetPaths()) != 1 {
			return fmt.Errorf("invalid field path: '*' must not be used with other paths")
		}
		return nil
	}
	md0 := m.ProtoReflect().Descriptor()
	for _, path := range fm.GetPaths() {
		md := md0
		if !rangeFields(path, func(field string) bool {
			// Search the field within the message.
			if md == nil {
				return false // not within a message
			}
			fd := md.Fields().ByName(protoreflect.Name(field))
			// The real field name of a group is the message name.
			if fd == nil {
				gd := md.Fields().ByName(protoreflect.Name(strings.ToLower(field)))
				if gd != nil && gd.Kind() == protoreflect.GroupKind && string(gd.Message().Name()) == field {
					fd = gd
				}
			} else if fd.Kind() == protoreflect.GroupKind && string(fd.Message().Name()) != field {
				fd = nil
			}
			if fd == nil {
				return false // message has does not have this field
			}
			// Identify the next message to search within.
			md = fd.Message() // may be nil
			if fd.IsMap() {
				md = fd.MapValue().Message() // may be nil
			}
			return true
		}) {
			return fmt.Errorf("invalid field path: %s", path)
		}
	}
	return nil
}

func stringsContain(str string, ss []string) bool {
	for _, s := range ss {
		if s == str {
			return true
		}
	}
	return false
}

func rangeFields(path string, f func(field string) bool) bool {
	for {
		var field string
		if i := strings.IndexByte(path, '.'); i >= 0 {
			field, path = path[:i], path[i:]
		} else {
			field, path = path, ""
		}
		if !f(field) {
			return false
		}
		if len(path) == 0 {
			return true
		}
		path = strings.TrimPrefix(path, ".")
	}
}
