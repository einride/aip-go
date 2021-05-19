package resourcename

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParent(t *testing.T) {
	assert.Equal(t, "", Parent(""))
	assert.Equal(t, "", Parent("foo"))
	assert.Equal(t, "foo", Parent("foo/bar"))
	assert.Equal(t, "foo/bar", Parent("foo/bar/baz"))
	assert.Equal(t, "", Parent("//example.com/foo"))
	assert.Equal(t, "foo", Parent("//example.com/foo/bar"))
	assert.Equal(t, "foo/bar", Parent("//example.com/foo/bar/baz"))
	assert.Equal(t, "foo/{foo}/bar", Parent("foo/{foo}/bar/{bar}"))
}
