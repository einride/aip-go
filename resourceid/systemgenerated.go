package resourceid

import (
	"encoding/base32"

	"github.com/google/uuid"
)

//nolint: gochecknoglobals
var base32Encoding = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

// NewSystemGenerated returns a new system-generated resource ID.
func NewSystemGenerated() string {
	return uuid.New().String()
}

// NewSystemGenerated returns a new system-generated resource ID encoded as base32 lowercase.
func NewSystemGeneratedBase32() string {
	id := uuid.New()
	return base32Encoding.EncodeToString(id[:])
}
