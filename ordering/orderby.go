package ordering

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// OrderBy represents an ordering directive.
type OrderBy struct {
	// Fields are the fields to order by.
	Fields []Field
}

// IsValidForMessage reports whether all the ordering paths are syntactically valid and
// refer to known fields in the specified message type.
func (o OrderBy) IsValidForMessage(m proto.Message) bool {
	mask := fieldmaskpb.FieldMask{
		Paths: make([]string, 0, len(o.Fields)),
	}
	for _, field := range o.Fields {
		mask.Paths = append(mask.Paths, field.Path)
	}
	return mask.IsValid(m)
}

// ValidateForPaths validates that the ordering paths are syntactically valid and refer to one of the provided paths.
func (o OrderBy) ValidateForPaths(paths ...string) error {
FieldLoop:
	for _, field := range o.Fields {
		// Assumption that len(paths) is short enough that O(n^2) is not a problem.
		for _, path := range paths {
			if field.Path == path {
				continue FieldLoop
			}
		}
		return fmt.Errorf("invalid field path: %s", field.Path)
	}
	return nil
}

// Field represents a single ordering field.
type Field struct {
	// Path is the path of the field, including subfields.
	Path string
	// Desc indicates if the ordering of the field is descending.
	Desc bool
}

// SubFields returns the individual subfields of the field path, including the top-level subfield.
//
// Subfields are specified with a . character, such as foo.bar or address.street.
func (f Field) SubFields() []string {
	if f.Path == "" {
		return nil
	}
	return strings.Split(f.Path, ".")
}

// UnmarshalString sets o from the provided ordering string. .
func (o *OrderBy) UnmarshalString(s string) error {
	o.Fields = o.Fields[:0]
	if s == "" { // fast path for no ordering
		return nil
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != ' ' && r != ',' && r != '.' {
			return fmt.Errorf("unmarshal order by '%s': invalid character %s", s, strconv.QuoteRune(r))
		}
	}
	fields := strings.Split(s, ",")
	o.Fields = make([]Field, 0, len(fields))
	for _, field := range fields {
		parts := strings.Fields(field)
		switch len(parts) {
		case 1: // default ordering (ascending)
			o.Fields = append(o.Fields, Field{Path: parts[0]})
		case 2: // specific ordering
			order := parts[1]
			var desc bool
			switch order {
			case "asc":
				desc = false
			case "desc":
				desc = true
			default: // parse error
				return fmt.Errorf("unmarshal order by '%s': invalid format", s)
			}
			o.Fields = append(o.Fields, Field{Path: parts[0], Desc: desc})
		case 0:
			fallthrough
		default:
			return fmt.Errorf("unmarshal order by '%s': invalid format", s)
		}
	}
	return nil
}
