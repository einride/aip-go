package aipreflect

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestResourceTypeName_Validate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		errorContains string
	}{
		{
			name: "pubsub.googleapis.com/Topic",
		},
		{
			name:          "pubsub/Topic",
			errorContains: "service name: must be a valid domain name",
		},
		{
			name:          "pubsub.googleapis.com/topic",
			errorContains: "type: must start with an upper-case letter",
		},
		{
			name:          "pubsub.googleapis.com/Topic_2",
			errorContains: "type: must be UpperCamelCase",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.errorContains != "" {
				assert.ErrorContains(t, ResourceType(tt.name).Validate(), tt.errorContains)
			} else {
				assert.NilError(t, ResourceType(tt.name).Validate())
			}
		})
	}
}
