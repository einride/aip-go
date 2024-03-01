package validation

import (
	"testing"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestError_NewError(t *testing.T) {
	t.Parallel()
	t.Run("panics on empty field violations", func(t *testing.T) {
		t.Parallel()
		assert.Assert(t, cmp.Panics(func() {
			_ = NewError(nil)
		}))
	})
}

func TestError_Error(t *testing.T) {
	t.Parallel()
	err := NewError([]*errdetails.BadRequest_FieldViolation{
		{Field: "foo.bar", Description: "test"},
		{Field: "baz", Description: "test2"},
	})
	assert.Error(t, err, `field violation on multiple fields:
 | foo.bar: test
 | baz: test2`)
}

func TestError_GRPCStatus(t *testing.T) {
	t.Parallel()
	expected := &errdetails.BadRequest{
		FieldViolations: []*errdetails.BadRequest_FieldViolation{
			{Field: "foo.bar", Description: "test"},
			{Field: "baz", Description: "test2"},
		},
	}
	s := status.Convert(NewError(expected.GetFieldViolations()))
	assert.Equal(t, codes.InvalidArgument, s.Code())
	assert.Equal(t, "invalid fields: foo.bar, baz", s.Message())
	details := s.Details()
	assert.Assert(t, len(details) == 1)
	actual, ok := details[0].(*errdetails.BadRequest)
	assert.Assert(t, ok)
	assert.DeepEqual(t, expected, actual, protocmp.Transform())
}
