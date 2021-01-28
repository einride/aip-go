package pagination

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestEncodePageTokenStruct(t *testing.T) {
	t.Parallel()
	const expected = "Kv-BAwEBCXBhZ2VUb2tlbgH_ggABAgEDSW50AQQAAQZTdHJpbmcBDAAAAAr_ggFUAQNmb28A"
	type pageToken struct {
		Int    int
		String string
	}
	token := pageToken{
		Int:    42,
		String: "foo",
	}
	assert.Equal(t, expected, EncodePageTokenStruct(&token))
}

func TestDecodePageTokenStruct(t *testing.T) {
	t.Parallel()
	type pageToken struct {
		Int    int
		String string
	}
	var actual pageToken
	const input = "Kv-BAwEBCXBhZ2VUb2tlbgH_ggABAgEDSW50AQQAAQZTdHJpbmcBDAAAAAr_ggFUAQNmb28A"
	assert.NilError(t, DecodePageTokenStruct(input, &actual))
	expected := pageToken{
		Int:    42,
		String: "foo",
	}
	assert.Equal(t, expected, actual)
}
