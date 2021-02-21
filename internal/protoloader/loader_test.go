package protoloader

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestLoadFilesFromGoPackage(t *testing.T) {
	t.Parallel()
	files, err := LoadFilesFromGoPackage("go.einride.tech/aip/examples/proto/gen/einride/example/freight/v1")
	assert.NilError(t, err)
	assert.Assert(t, files.NumFiles() > 0)
}
