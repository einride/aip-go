package resourceid

import (
	"fmt"
	"unicode"
)

// ValidateUserSettable validates a user-settable resource ID.
//
// From https://google.aip.dev/122#resource-id-segments:
//
// User-settable resource IDs should conform to RFC-1034; which restricts to letters, numbers, and hyphen,
// with the first character a letter, the last a letter or a number, and a 63 character maximum.
// Additionally, user-settable resource IDs should restrict letters to lower-case.
//
// User-settable IDs should not be permitted to be a UUID (or any value that syntactically appears to be a UUID).
//
// See also: https://google.aip.dev/133#user-specified-ids
func ValidateUserSettable(id string) error {
	if len(id) < 1 || 63 < len(id) {
		return fmt.Errorf("user-settable ID must be between 1 and 63 characters")
	}
	if !unicode.IsLetter(rune(id[0])) {
		return fmt.Errorf("user-settable ID must begin with a letter")
	}
	if id[len(id)-1] == '-' {
		return fmt.Errorf("user-settable ID must end with a letter or number")
	}
	for position, character := range id {
		switch character {
		case
			// numbers
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			// hyphen
			'-',
			// lower-case
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		default:
			return fmt.Errorf(
				"user-settable ID must only contain lowercase, numbers and hyphens (got: '%c' in position %d)",
				character,
				position,
			)
		}
	}
	return nil
}
