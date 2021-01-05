package aipreflect

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestResourceTypeDescriptor(t *testing.T) {
	t.Parallel()
	t.Run("can be used as map key", func(t *testing.T) {
		t.Parallel()
		m := map[ResourceTypeDescriptor]string{}
		m[ResourceTypeDescriptor{ServiceName: "pubsub.googleapis.com", Type: "Topic"}] = "test"
		assert.Equal(t, "test", m[ResourceTypeDescriptor{ServiceName: "pubsub.googleapis.com", Type: "Topic"}])
	})
}

func TestResourceTypeDescriptor_String(t *testing.T) {
	t.Parallel()
	const expected = "pubsub.googleapis.com/Topic"
	actual := ResourceTypeDescriptor{ServiceName: "pubsub.googleapis.com", Type: "Topic"}.String()
	assert.Equal(t, expected, actual)
}

func TestResourceTypeDescriptor_Validate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		test          string
		desc          ResourceTypeDescriptor
		errorContains string
	}{
		{
			test: "valid",
			desc: ResourceTypeDescriptor{ServiceName: "pubsub.googleapis.com", Type: "Topic"},
		},
		{
			test:          "invalid service name",
			desc:          ResourceTypeDescriptor{ServiceName: "pubsub", Type: "Topic"},
			errorContains: "service name: must be a valid domain name",
		},
		{
			test:          "lower-case type",
			desc:          ResourceTypeDescriptor{ServiceName: "pubsub.googleapis.com", Type: "topic"},
			errorContains: "type: must start with an upper-case letter",
		},
		{
			test:          "snake-case type",
			desc:          ResourceTypeDescriptor{ServiceName: "pubsub.googleapis.com", Type: "Topic_2"},
			errorContains: "type: must be UpperCamelCase",
		},
	} {
		tt := tt
		t.Run(tt.test, func(t *testing.T) {
			t.Parallel()
			if tt.errorContains != "" {
				assert.ErrorContains(t, tt.desc.Validate(), tt.errorContains)
			} else {
				assert.NilError(t, tt.desc.Validate())
			}
		})
	}
}
