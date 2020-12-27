package resourceid

import (
	"regexp"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestNewSystemGenerated(t *testing.T) {
	t.Parallel()
	const uuidV4Regexp = `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	assert.Assert(t, cmp.Regexp(regexp.MustCompile(uuidV4Regexp), NewSystemGenerated()))
}
