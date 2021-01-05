package aipspansql

import (
	"testing"

	"cloud.google.com/go/spanner/spansql"
	"go.einride.tech/aip/ordering"
	"gotest.tools/v3/assert"
)

func TestOrder(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		orderBy  ordering.OrderBy
		expected []spansql.Order
	}{
		{
			name:     "empty",
			orderBy:  ordering.OrderBy{},
			expected: nil,
		},

		{
			name: "single field",
			orderBy: ordering.OrderBy{
				Fields: []ordering.Field{
					{Path: "foo"},
				},
			},
			expected: []spansql.Order{
				{Expr: spansql.ID("foo")},
			},
		},

		{
			name: "single field, with subfields",
			orderBy: ordering.OrderBy{
				Fields: []ordering.Field{
					{Path: "foo.bar"},
				},
			},
			expected: []spansql.Order{
				{Expr: spansql.PathExp{spansql.ID("foo"), spansql.ID("bar")}},
			},
		},

		{
			name: "single field, desc",
			orderBy: ordering.OrderBy{
				Fields: []ordering.Field{
					{Path: "foo", Desc: true},
				},
			},
			expected: []spansql.Order{
				{Expr: spansql.ID("foo"), Desc: true},
			},
		},

		{
			name: "multiple fields, with subfields, desc",
			orderBy: ordering.OrderBy{
				Fields: []ordering.Field{
					{Path: "foo"},
					{Path: "bar.baz", Desc: true},
				},
			},
			expected: []spansql.Order{
				{Expr: spansql.ID("foo")},
				{Expr: spansql.PathExp{spansql.ID("bar"), spansql.ID("baz")}, Desc: true},
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.DeepEqual(t, tt.expected, Order(tt.orderBy))
		})
	}
}
