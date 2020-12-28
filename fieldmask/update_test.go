package fieldmask

import (
	"testing"

	"google.golang.org/genproto/googleapis/example/library/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestUpdateTopLevelFields(t *testing.T) {
	t.Parallel()
	t.Run("should panic on different src and dst", func(t *testing.T) {
		t.Parallel()
		assert.Assert(t, cmp.Panics(func() {
			UpdateTopLevelFields(&fieldmaskpb.FieldMask{}, &library.Book{}, &library.Shelf{})
		}))
	})

	t.Run("test cases", func(t *testing.T) {
		t.Parallel()
		for _, tt := range []struct {
			name     string
			mask     *fieldmaskpb.FieldMask
			src      proto.Message
			dst      proto.Message
			expected proto.Message
		}{
			{
				name: "full replacement",
				mask: &fieldmaskpb.FieldMask{
					Paths: []string{"*"},
				},
				src: &library.Book{
					Name:   "src-name",
					Author: "src-author",
					Title:  "src-title",
					Read:   false,
				},
				dst: &library.Book{
					Name:   "dst-name",
					Author: "dst-author",
					Title:  "dst-title",
					Read:   true,
				},
				expected: &library.Book{
					Name:   "src-name",
					Author: "src-author",
					Title:  "src-title",
					Read:   false,
				},
			},

			{
				name: "no field mask replaces non-zero fields",
				mask: nil,
				src: &library.Book{
					Name:   "src-name",
					Author: "src-author",
				},
				dst: &library.Book{
					Name:   "dst-name",
					Author: "dst-author",
					Title:  "dst-title",
					Read:   true,
				},
				expected: &library.Book{
					Name:   "src-name",
					Author: "src-author",
					Title:  "dst-title",
					Read:   true,
				},
			},

			{
				name: "with partial field mask",
				mask: &fieldmaskpb.FieldMask{
					Paths: []string{"name", "author", "read"},
				},
				src: &library.Book{
					Name:   "src-name",
					Author: "src-author",
					Title:  "src-title",
					Read:   false,
				},
				dst: &library.Book{
					Name:   "dst-name",
					Author: "dst-author",
					Title:  "dst-title",
					Read:   true,
				},
				expected: &library.Book{
					Name:   "src-name",
					Author: "src-author",
					Title:  "dst-title",
					Read:   false,
				},
			},

			{
				name: "with non-top-level path",
				mask: &fieldmaskpb.FieldMask{
					Paths: []string{"name", "book.name"},
				},
				src: &library.CreateBookRequest{
					Name: "src-shelf",
					Book: &library.Book{
						Name:   "src-name",
						Author: "src-author",
						Title:  "src-title",
						Read:   false,
					},
				},
				dst: &library.CreateBookRequest{
					Name: "dst-shelf",
					Book: &library.Book{
						Name:   "dst-name",
						Author: "dst-author",
						Title:  "dst-title",
						Read:   true,
					},
				},
				expected: &library.CreateBookRequest{
					Name: "src-shelf",
					Book: &library.Book{
						Name:   "dst-name",
						Author: "dst-author",
						Title:  "dst-title",
						Read:   true,
					},
				},
			},

			{
				name: "with unknown path",
				mask: &fieldmaskpb.FieldMask{
					Paths: []string{"name", "foo"},
				},
				src: &library.Book{
					Name:   "src-name",
					Author: "src-author",
					Title:  "src-title",
					Read:   false,
				},
				dst: &library.Book{
					Name:   "dst-name",
					Author: "dst-author",
					Title:  "dst-title",
					Read:   true,
				},
				expected: &library.Book{
					Name:   "src-name",
					Author: "dst-author",
					Title:  "dst-title",
					Read:   true,
				},
			},
		} {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				UpdateTopLevelFields(tt.mask, tt.dst, tt.src)
				assert.DeepEqual(t, tt.expected, tt.dst, protocmp.Transform())
			})
		}
	})
}
