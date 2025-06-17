package pagination

import (
	"testing"

	"gotest.tools/v3/assert"
)

func Test_PageTokenStruct(t *testing.T) {
	t.Parallel()
	type pageToken struct {
		Int    int
		String string
	}
	for _, tt := range []struct {
		name string
		in   pageToken
	}{
		{
			name: "all set",
			in: pageToken{
				Int:    42,
				String: "foo",
			},
		},
		{
			name: "default value",
			in: pageToken{
				String: "foo",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			str := EncodePageTokenStruct(tt.in)
			var out pageToken
			assert.NilError(t, DecodePageTokenStruct(str, &out), str)
			assert.Equal(t, tt.in, out)
		})
	}
}
