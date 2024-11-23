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

func TestNewSystemGeneratedBase32(t *testing.T) {
	t.Parallel()
	const base32Regexp = `^[a-z2-7]{26}$`
	assert.Assert(t, cmp.Regexp(regexp.MustCompile(base32Regexp), NewSystemGeneratedBase32()))
}

func TestNewSystemGeneratedBase32FollowingAIP122(t *testing.T) {
	t.Parallel()
	for i := 0; i < 100; i++ {
		// https://google.aip.dev/122#resource-id-segments
		const base32Regexp = `^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$`
		assert.Assert(t, cmp.Regexp(regexp.MustCompile(base32Regexp), NewSystemGeneratedBase32()))
	}
}
