package resourceid

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestValidateUserSettable(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		id            string
		errorContains string
	}{
		{id: "abcd"},
		{id: "abcd-efgh-1234"},
		{id: "", errorContains: "must be between 1 and 63 characters"},
		{id: strings.Repeat("a", 64), errorContains: "must be between 1 and 63 characters"},
		{id: "-abc", errorContains: "must begin with a letter"},
		{id: "abc-", errorContains: "must end with a letter or number"},
		{id: "123-abc", errorContains: "must begin with a letter"},
		{id: "daf1cb3e-f33b-43f1-81cc-e65fda51efa5", errorContains: "must not be a valid UUIDv4"},
		{id: "abcd/efgh", errorContains: "must only contain lowercase, numbers and hyphens"},
	} {
		tt := tt
		t.Run(tt.id, func(t *testing.T) {
			t.Parallel()
			err := ValidateUserSettable(tt.id)
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
