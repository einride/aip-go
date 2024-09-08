package toplevelsingleton

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestConfigResourceName(t *testing.T) {
	t.Run("good", func(t *testing.T) {
		const pattern = "config"
		var name ConfigResourceName
		err := name.UnmarshalString(pattern)
		assert.NilError(t, err)

		marshalled, err := name.MarshalString()
		assert.NilError(t, err)
		assert.Equal(t, marshalled, pattern)
	})

	t.Run("invalid", func(t *testing.T) {
		var name ConfigResourceName
		err := name.UnmarshalString("other")
		t.Log(err)
		assert.Error(t, err, "parse resource name 'other' with pattern 'config': segment config: got other")
	})
}
