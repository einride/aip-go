package resourcetype

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestName(t *testing.T) {
	t.Parallel()
	t.Run("can be used as map key", func(t *testing.T) {
		t.Parallel()
		m := map[Name]string{}
		m[Name{ServiceName: "pubsub.googleapis.com", Type: "Topic"}] = "test"
		assert.Equal(t, "test", m[Name{ServiceName: "pubsub.googleapis.com", Type: "Topic"}])
	})
}

func TestName_String(t *testing.T) {
	t.Parallel()
	const expected = "pubsub.googleapis.com/Topic"
	actual := Name{ServiceName: "pubsub.googleapis.com", Type: "Topic"}.String()
	assert.Equal(t, expected, actual)
}

func TestName_Validate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		test          string
		name          Name
		errorContains string
	}{
		{
			test: "valid",
			name: Name{ServiceName: "pubsub.googleapis.com", Type: "Topic"},
		},
		{
			test:          "invalid service name",
			name:          Name{ServiceName: "pubsub", Type: "Topic"},
			errorContains: "service name: must be a valid domain name",
		},
		{
			test:          "lower-case type",
			name:          Name{ServiceName: "pubsub.googleapis.com", Type: "topic"},
			errorContains: "type: must start with an upper-case letter",
		},
		{
			test:          "snake-case type",
			name:          Name{ServiceName: "pubsub.googleapis.com", Type: "Topic_2"},
			errorContains: "type: must be UpperCamelCase",
		},
	} {
		tt := tt
		t.Run(tt.test, func(t *testing.T) {
			t.Parallel()
			err := tt.name.Validate()
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
