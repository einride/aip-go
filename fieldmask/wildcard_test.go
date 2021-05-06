package fieldmask

import (
	"testing"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gotest.tools/v3/assert"
)

func TestIsFullReplacement(t *testing.T) {
	t.Parallel()
	assert.Assert(t, IsFullReplacement(&fieldmaskpb.FieldMask{Paths: []string{WildcardPath}}))
	assert.Assert(t, !IsFullReplacement(&fieldmaskpb.FieldMask{Paths: []string{WildcardPath, "foo"}}))
	assert.Assert(t, !IsFullReplacement(&fieldmaskpb.FieldMask{Paths: []string{"foo"}}))
	assert.Assert(t, !IsFullReplacement(nil))
}
