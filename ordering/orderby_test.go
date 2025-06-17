package ordering

import (
	"testing"

	"google.golang.org/genproto/googleapis/example/library/v1"
	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"
)

func TestOrderBy_UnmarshalString(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		orderBy       string
		expected      OrderBy
		errorContains string
	}{
		{
			orderBy:  "",
			expected: OrderBy{},
		},

		{
			orderBy: "foo desc, bar",
			expected: OrderBy{
				Fields: []Field{
					{Path: "foo", Desc: true},
					{Path: "bar"},
				},
			},
		},

		{
			orderBy: "foo.bar",
			expected: OrderBy{
				Fields: []Field{
					{Path: "foo.bar"},
				},
			},
		},

		{
			orderBy: " foo , bar desc ",
			expected: OrderBy{
				Fields: []Field{
					{Path: "foo"},
					{Path: "bar", Desc: true},
				},
			},
		},

		{orderBy: "foo,", errorContains: "invalid format"},
		{orderBy: ",", errorContains: "invalid "},
		{orderBy: ",foo", errorContains: "invalid format"},
		{orderBy: "foo/bar", errorContains: "invalid character '/'"},
		{orderBy: "foo bar", errorContains: "invalid format"},
	} {
		t.Run(tt.orderBy, func(t *testing.T) {
			t.Parallel()
			var actual OrderBy
			err := actual.UnmarshalString(tt.orderBy)
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, tt.expected, actual)
			}
		})
	}
}

func TestOrderBy_ValidateForPaths(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		orderBy       OrderBy
		paths         []string
		errorContains string
	}{
		{
			name:    "valid empty",
			orderBy: OrderBy{},
			paths:   []string{},
		},

		{
			name: "valid single",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "name"},
					{Path: "author"},
				},
			},
			paths: []string{"name", "author", "read"},
		},

		{
			name: "invalid single",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "name"},
					{Path: "foo"},
				},
			},
			paths:         []string{"name", "author", "read"},
			errorContains: "invalid field path: foo",
		},

		{
			name: "valid nested",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "name"},
					{Path: "book.name"},
				},
			},
			paths: []string{"name", "book.name", "book.author", "book.read"},
		},

		{
			name: "invalid nested",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "name"},
					{Path: "book.foo"},
				},
			},
			paths:         []string{"name", "book.name", "book.author", "book.read"},
			errorContains: "invalid field path: book.foo",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.errorContains != "" {
				assert.ErrorContains(t, tt.orderBy.ValidateForPaths(tt.paths...), tt.errorContains)
			} else {
				assert.NilError(t, tt.orderBy.ValidateForPaths(tt.paths...))
			}
		})
	}
}

func TestOrderBy_ValidateForMessage(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		orderBy       OrderBy
		message       proto.Message
		errorContains string
	}{
		{
			name:    "valid empty",
			orderBy: OrderBy{},
			message: &library.Book{},
		},

		{
			name: "valid single",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "name"},
					{Path: "author"},
				},
			},
			message: &library.Book{},
		},

		{
			name: "invalid single",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "name"},
					{Path: "foo"},
				},
			},
			message:       &library.Book{},
			errorContains: "invalid field path: foo",
		},

		{
			name: "valid nested",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "parent"},
					{Path: "book.name"},
				},
			},
			message: &library.CreateBookRequest{},
		},

		{
			name: "invalid nested",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "parent"},
					{Path: "book.foo"},
				},
			},
			message:       &library.CreateBookRequest{},
			errorContains: "invalid field path: book.foo",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.errorContains != "" {
				assert.ErrorContains(t, tt.orderBy.ValidateForMessage(tt.message), tt.errorContains)
			} else {
				assert.NilError(t, tt.orderBy.ValidateForMessage(tt.message))
			}
		})
	}
}

func TestField_SubFields(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		field    Field
		expected []string
	}{
		{
			name:     "empty",
			field:    Field{},
			expected: nil,
		},

		{
			name:     "single",
			field:    Field{Path: "foo"},
			expected: []string{"foo"},
		},

		{
			name:     "multiple",
			field:    Field{Path: "foo.bar"},
			expected: []string{"foo", "bar"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.DeepEqual(t, tt.expected, tt.field.SubFields())
		})
	}
}
