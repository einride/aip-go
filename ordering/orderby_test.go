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
		tt := tt
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

func TestOrderBy_IsValidForMessage(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		orderBy OrderBy
		message proto.Message
		isValid bool
	}{
		{
			name:    "valid empty",
			orderBy: OrderBy{},
			message: &library.Book{},
			isValid: true,
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
			isValid: true,
		},

		{
			name: "invalid single",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "name"},
					{Path: "foo"},
				},
			},
			message: &library.Book{},
			isValid: false,
		},

		{
			name: "valid nested",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "name"},
					{Path: "book.name"},
				},
			},
			message: &library.CreateBookRequest{},
			isValid: true,
		},

		{
			name: "invalid nested",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "name"},
					{Path: "book.foo"},
				},
			},
			message: &library.CreateBookRequest{},
			isValid: false,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.isValid, tt.orderBy.IsValidForMessage(tt.message))
		})
	}
}

func TestOrderBy_IsValidForPaths(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		orderBy OrderBy
		paths   []string
		isValid bool
	}{
		{
			name:    "valid empty",
			orderBy: OrderBy{},
			paths:   []string{},
			isValid: true,
		},

		{
			name: "valid single",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "name"},
					{Path: "author"},
				},
			},
			paths:   []string{"name", "author", "read"},
			isValid: true,
		},

		{
			name: "invalid single",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "name"},
					{Path: "foo"},
				},
			},
			paths:   []string{"name", "author", "read"},
			isValid: false,
		},

		{
			name: "valid nested",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "name"},
					{Path: "book.name"},
				},
			},
			paths:   []string{"name", "book.name", "book.author", "book.read"},
			isValid: true,
		},

		{
			name: "invalid nested",
			orderBy: OrderBy{
				Fields: []Field{
					{Path: "name"},
					{Path: "book.foo"},
				},
			},
			paths:   []string{"name", "book.name", "book.author", "book.read"},
			isValid: false,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.isValid, tt.orderBy.IsValidForPaths(tt.paths...))
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.DeepEqual(t, tt.expected, tt.field.SubFields())
		})
	}
}
