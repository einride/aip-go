package aipreflect

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGrammaticalName_Validate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		errorContains string
	}{
		{name: "", errorContains: "must be non-empty"},
		{name: "users"},
		{name: "userEvents"},
		{name: "UserEvents", errorContains: "must be lowerCamelCase"},
		{name: "user events", errorContains: "contains forbidden character ' '"},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := GrammaticalName(tt.name).Validate()
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
