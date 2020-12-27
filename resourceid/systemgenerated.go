package resourceid

import "github.com/google/uuid"

// NewSystemGenerated returns a new system-generated resource ID.
func NewSystemGenerated() string {
	return uuid.New().String()
}
