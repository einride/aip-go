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

func TestUpdate(t *testing.T) {
	t.Parallel()
	t.Run("should panic on different src and dst", func(t *testing.T) {
		t.Parallel()
		assert.Assert(t, cmp.Panics(func() {
			Update(&fieldmaskpb.FieldMask{}, &library.Book{}, &library.Shelf{})
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
				name: "full replacement: nested",
				mask: &fieldmaskpb.FieldMask{
					Paths: []string{"*"},
				},
				src: &library.CreateBookRequest{
					Name: "src-name",
					Book: &library.Book{
						Author: "src-author",
						Read:   false,
					},
				},
				dst: &library.CreateBookRequest{
					Name: "dst-name",
					Book: &library.Book{
						Author: "dst-author",
						Title:  "dst-title",
						Read:   true,
					},
				},
				expected: &library.CreateBookRequest{
					Name: "src-name",
					Book: &library.Book{
						Author: "src-author",
						Read:   false,
					},
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
				name: "no field mask replaces non-zero fields: nested",
				mask: nil,
				src: &library.CreateBookRequest{
					Book: &library.Book{
						Name:   "src-name",
						Author: "src-author",
					},
				},
				dst: &library.CreateBookRequest{
					Book: &library.Book{
						Name:   "dst-name",
						Author: "dst-author",
						Title:  "dst-title",
						Read:   true,
					},
				},
				expected: &library.CreateBookRequest{
					Book: &library.Book{
						Name:   "src-name",
						Author: "src-author",
						Title:  "dst-title",
						Read:   true,
					},
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
						Name:   "src-name",
						Author: "dst-author",
						Title:  "dst-title",
						Read:   true,
					},
				},
			},

			{
				name: "repeated field",
				mask: &fieldmaskpb.FieldMask{
					Paths: []string{"books"},
				},
				src: &library.ListBooksResponse{
					NextPageToken: "xxx",
					Books: []*library.Book{
						{
							Name:   "src-name",
							Author: "src-author",
							Title:  "src-title",
							Read:   false,
						},
						{
							Name:   "src-name-2",
							Author: "src-author-2",
							Title:  "src-title-2",
							Read:   false,
						},
					},
				},
				dst: &library.ListBooksResponse{
					NextPageToken: "yyy",
					Books: []*library.Book{
						{
							Name:   "dst-name",
							Author: "dst-author",
							Title:  "dst-title",
							Read:   false,
						},
					},
				},
				expected: &library.ListBooksResponse{
					NextPageToken: "yyy",
					Books: []*library.Book{
						{
							Name:   "src-name",
							Author: "src-author",
							Title:  "src-title",
							Read:   false,
						},
						{
							Name:   "src-name-2",
							Author: "src-author-2",
							Title:  "src-title-2",
							Read:   false,
						},
					},
				},
			},

			{
				name: "repeated field nested path",
				mask: &fieldmaskpb.FieldMask{
					Paths: []string{"books.name"},
				},
				src: &library.ListBooksResponse{
					NextPageToken: "xxx",
					Books: []*library.Book{
						{
							Name:   "src-name",
							Author: "src-author",
							Title:  "src-title",
							Read:   false,
						},
						{
							Name:   "src-name-2",
							Author: "src-author-2",
							Title:  "src-title-2",
							Read:   false,
						},
					},
				},
				dst: &library.ListBooksResponse{
					NextPageToken: "yyy",
					Books: []*library.Book{
						{
							Name:   "dst-name",
							Author: "dst-author",
							Title:  "dst-title",
							Read:   false,
						},
					},
				},
				expected: &library.ListBooksResponse{
					NextPageToken: "yyy",
					Books: []*library.Book{
						{
							Name:   "dst-name",
							Author: "dst-author",
							Title:  "dst-title",
							Read:   false,
						},
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
				Update(tt.mask, tt.dst, tt.src)
				assert.DeepEqual(t, tt.expected, tt.dst, protocmp.Transform())
			})
		}
	})
}
