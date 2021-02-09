package fieldmask

import (
	"testing"

	"google.golang.org/genproto/googleapis/example/library/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gotest.tools/v3/assert"
)

func TestValidate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		fieldMask     *fieldmaskpb.FieldMask
		message       proto.Message
		errorContains string
	}{
		{
			name:    "valid nil",
			message: &library.Book{},
		},
		{
			name:      "valid *",
			fieldMask: &fieldmaskpb.FieldMask{Paths: []string{"*"}},
			message:   &library.Book{},
		},
		{
			name:          "invalid *",
			fieldMask:     &fieldmaskpb.FieldMask{Paths: []string{"*", "author"}},
			message:       &library.Book{},
			errorContains: "invalid field path: '*' must not be used with other paths",
		},
		{
			name:      "valid empty",
			fieldMask: &fieldmaskpb.FieldMask{},
			message:   &library.Book{},
		},

		{
			name: "valid single",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"name", "author"},
			},
			message: &library.Book{},
		},

		{
			name: "invalid single",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"name", "foo"},
			},
			message:       &library.Book{},
			errorContains: "invalid field path: foo",
		},

		{
			name: "valid nested",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"name", "book.name"},
			},
			message: &library.CreateBookRequest{},
		},

		{
			name: "invalid nested",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"name", "book.foo"},
			},
			message:       &library.CreateBookRequest{},
			errorContains: "invalid field path: book.foo",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.errorContains != "" {
				assert.ErrorContains(t, Validate(tt.fieldMask, tt.message), tt.errorContains)
			} else {
				assert.NilError(t, Validate(tt.fieldMask, tt.message))
			}
		})
	}
}
