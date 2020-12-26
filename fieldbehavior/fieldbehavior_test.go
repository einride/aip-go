package fieldbehavior

import (
	"testing"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/genproto/googleapis/example/library/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestValidateRequiredFields(t *testing.T) {
	t.Parallel()
	assert.NilError(t, ValidateRequiredFields(&library.Book{Name: "testbook"}))
	assert.Error(t, ValidateRequiredFields(&library.Book{}), "missing required field: name")
}

func TestCopyFields(t *testing.T) {
	t.Parallel()
	t.Run("different types", func(t *testing.T) {
		t.Parallel()
		assert.Assert(t, cmp.Panics(func() {
			CopyFields(&library.Book{}, &library.Shelf{}, annotations.FieldBehavior_REQUIRED)
		}))
	})
}

func TestValidateRequiredFieldsWithMask(t *testing.T) {
	t.Parallel()
	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		assert.NilError(t, ValidateRequiredFieldsWithMask(&library.Book{Name: "testbook"}, nil))
	})
	t.Run("missing field", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRequiredFieldsWithMask(&library.Book{}, nil), "missing required field: name")
	})
	t.Run("missing but not in mask", func(t *testing.T) {
		t.Parallel()
		assert.NilError(
			t,
			ValidateRequiredFieldsWithMask(&library.Book{}, &fieldmaskpb.FieldMask{Paths: []string{"author"}}),
		)
	})
	t.Run("missing nested", func(t *testing.T) {
		t.Parallel()
		assert.Error(
			t,
			ValidateRequiredFieldsWithMask(
				&library.UpdateBookRequest{
					Name: "testname",
					Book: &library.Book{},
				},
				nil,
			),
			"missing required field: book.name",
		)
	})
	t.Run("missing nested not in mask", func(t *testing.T) {
		t.Parallel()
		assert.NilError(
			t,
			ValidateRequiredFieldsWithMask(
				&library.UpdateBookRequest{
					Book: &library.Book{},
				},
				&fieldmaskpb.FieldMask{Paths: []string{"book.author"}},
			),
		)
	})
}
