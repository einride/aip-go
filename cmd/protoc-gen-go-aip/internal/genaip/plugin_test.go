package genaip

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
)

func TestRun(t *testing.T) {
	t.Parallel()
	withCompilerPluginPath(t)
	for _, tt := range []struct {
		name string
	}{
		{name: "single"},
		{name: "originallysinglepattern"},
		{name: "multipattern"},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			protoc(t,
				"-I", "../../../../examples/proto/api-common-protos",
				"-I", "testdata",
				"--go-aip_out=testdata",
				"--go-aip_opt=module=go.einride.tech/aip/cmd/protoc-gen-go-aip/internal/genaip/testdata",
				filepath.Join("testdata", tt.name, "testdata.proto"),
			)
		})
	}
}

func protoc(t *testing.T, args ...string) {
	t.Helper()
	var stderr bytes.Buffer
	cmd := exec.Command("protoc", args...)
	cmd.Stderr = &stderr
	assert.NilError(t, cmd.Run(), stderr.String())
}

func goBuild(t *testing.T, args ...string) {
	t.Helper()
	var stderr bytes.Buffer
	cmd := exec.Command("go", append([]string{"build"}, args...)...)
	cmd.Stderr = &stderr
	assert.NilError(t, cmd.Run(), stderr.String())
}

func withCompilerPluginPath(t *testing.T) {
	t.Helper()
	tmpDir, err := ioutil.TempDir(".", "")
	assert.NilError(t, err)
	t.Cleanup(func() {
		assert.NilError(t, os.RemoveAll(tmpDir))
	})
	goBuild(t, "-o", filepath.Join(tmpDir, PluginName), "../../")
	beforePath := os.Getenv("PATH")
	t.Cleanup(func() {
		assert.NilError(t, os.Setenv("PATH", beforePath))
	})
	assert.NilError(t, os.Setenv("PATH", tmpDir+":"+beforePath))
	t.Log(os.Getenv("PATH"))
}
