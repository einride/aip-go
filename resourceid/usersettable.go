package resourceid

import (
	"fmt"

	"github.com/google/uuid"
)

// ValidateUserSettable validates a user-settable resource ID.
//
// From https://google.aip.dev/122#resource-id-segments:
//
// User-settable resource IDs should conform to RFC-1034,which restricts to letters, numbers, and hyphen, with a 63
// character maximum. Additionally, user-settable resource IDs should restrict letters to lower-case.
//
// User-settable IDs should not be permitted to be a UUID (or any value that syntactically appears to be a UUID).
//
// See also: https://google.aip.dev/133#user-specified-ids
func ValidateUserSettable(id string) error {
	if len(id) < 4 || 63 < len(id) {
		return fmt.Errorf("user-settable ID must be between 4 and 63 characters")
	}
	if id[0] == '-' {
		return fmt.Errorf("user-settable ID must not start with a hyphen")
	}
	if _, err := uuid.Parse(id); err == nil {
		return fmt.Errorf("user-settable ID must not be a valid UUIDv4")
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
