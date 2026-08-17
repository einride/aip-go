package sqflop

import (
	"testing"

	"github.com/ysomad/flop/filter"
	"github.com/ysomad/flop/internal/assert"
)

func testFilterSchema(t *testing.T) *filter.Schema {
	t.Helper()
	schema, err := filter.NewSchema([]filter.Field{
		filter.NewField("state", filter.Enum("pending", "paid"), false),
		filter.NewField("total", filter.TypeInt, false),
		filter.NewField("name", filter.TypeString, true),
		filter.NewField("deleted", filter.TypeBool, false),
		filter.NewField("tags", filter.TypeStringList, false),
	}, filter.Limits{})
	assert.NoError(t, err)
	return schema
}

func TestFilterCompiler_Where(t *testing.T) {
	t.Parallel()

	compiler, err := NewFilterCompiler(testFilterSchema(t), map[string]string{
		"state":   "o.state",
		"total":   "o.total",
		"name":    "o.name",
		"deleted": "o.deleted",
		"tags":    "o.tags",
	})
	assert.NoError(t, err)

	tests := []struct {
		name     string
		filter   string
		wantSQL  string
		wantArgs []any
		wantErr  assert.ErrorAssertionFunc
	}{
		{name: "empty", wantSQL: "SELECT id FROM orders o", wantErr: assert.NoError},
		{
			name: "has", filter: `tags:"blue"`,
			wantSQL:  "SELECT id FROM orders o WHERE $1 = ANY(o.tags)",
			wantArgs: []any{"blue"},
			wantErr:  assert.NoError,
		},
		{
			name: "negated has", filter: `NOT tags:"blue"`,
			wantSQL:  "SELECT id FROM orders o WHERE NOT ($1 = ANY(o.tags))",
			wantArgs: []any{"blue"},
			wantErr:  assert.NoError,
		},
		{
			name: "has inside a disjunction", filter: `tags:"blue" OR tags:"red"`,
			wantSQL:  "SELECT id FROM orders o WHERE ($1 = ANY(o.tags) OR $2 = ANY(o.tags))",
			wantArgs: []any{"blue", "red"},
			wantErr:  assert.NoError,
		},
		{
			name: "negation distributes over has", filter: `NOT (tags:"blue" OR total < 10)`,
			wantSQL:  "SELECT id FROM orders o WHERE (NOT ($1 = ANY(o.tags)) AND o.total >= $2)",
			wantArgs: []any{"blue", int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name: "comparisons", filter: "state = paid AND total >= 10",
			wantSQL:  "SELECT id FROM orders o WHERE (o.state = $1 AND o.total >= $2)",
			wantArgs: []any{"paid", int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name:     "not equal",
			filter:   "total != 10",
			wantSQL:  "SELECT id FROM orders o WHERE o.total <> $1",
			wantArgs: []any{int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name:     "less",
			filter:   "total < 10",
			wantSQL:  "SELECT id FROM orders o WHERE o.total < $1",
			wantArgs: []any{int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name:     "less or equal",
			filter:   "total <= 10",
			wantSQL:  "SELECT id FROM orders o WHERE o.total <= $1",
			wantArgs: []any{int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name:     "greater",
			filter:   "total > 10",
			wantSQL:  "SELECT id FROM orders o WHERE o.total > $1",
			wantArgs: []any{int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name: "null", filter: "name != null",
			wantSQL: "SELECT id FROM orders o WHERE o.name IS NOT NULL",
			wantErr: assert.NoError,
		},
		{
			name: "boundary wildcard", filter: `name = "book*"`,
			wantSQL: "SELECT id FROM orders o WHERE o.name LIKE $1", wantArgs: []any{"book%"},
			wantErr: assert.NoError,
		},
		{
			name: "negative boundary wildcard", filter: `name != "*book"`,
			wantSQL: "SELECT id FROM orders o WHERE o.name NOT LIKE $1", wantArgs: []any{"%book"},
			wantErr: assert.NoError,
		},
		{
			name: "text sequence", filter: `"blue" "chair"`,
			wantSQL:  "SELECT id FROM orders o WHERE ((o.name LIKE $1) AND (o.name LIKE $2))",
			wantArgs: []any{"%blue%", "%chair%"},
			wantErr:  assert.NoError,
		},
		{
			name: "logical or", filter: `"blue" OR "chair"`,
			wantSQL:  "SELECT id FROM orders o WHERE ((o.name LIKE $1) OR (o.name LIKE $2))",
			wantArgs: []any{"%blue%", "%chair%"},
			wantErr:  assert.NoError,
		},
		{
			name: "negation", filter: "NOT (state = paid OR total < 10)",
			wantSQL:  "SELECT id FROM orders o WHERE (o.state <> $1 AND o.total >= $2)",
			wantArgs: []any{"paid", int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name: "negated text", filter: `NOT "blue"`,
			wantSQL: "SELECT id FROM orders o WHERE (o.name NOT LIKE $1)", wantArgs: []any{"%blue%"},
			wantErr: assert.NoError,
		},
		{name: "invalid filter", filter: "missing = 1", wantErr: assert.Error},
		{name: "unsupported pattern", filter: `name = "*a_b*"`, wantErr: assert.Error},
		{name: "unsupported text pattern", filter: `"a%b"`, wantErr: assert.Error},
		{name: "unsupported left expression", filter: `"a%b" AND "ok"`, wantErr: assert.Error},
		{name: "unsupported right expression", filter: `"ok" AND "a%b"`, wantErr: assert.Error},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			filter, err := filter.Parse(test.filter)
			assert.NoError(t, err)

			gotSQL, gotArgs, gotErr := compiler.Where(psql.Select("id").From("orders o"), filter).ToSql()
			test.wantErr(t, gotErr)
			if gotErr != nil {
				return
			}

			assert.Equal(t, test.wantSQL, gotSQL)
			assert.Equal(t, test.wantArgs, gotArgs)
		})
	}
}

func TestFilterCompiler_ApplyDefaultColumns(t *testing.T) {
	t.Parallel()
	schema := testFilterSchema(t)
	unmapped, err := NewFilterCompiler(schema, nil)
	assert.NoError(t, err)
	partial, err := NewFilterCompiler(schema, map[string]string{"total": "o.total"})
	assert.NoError(t, err)
	tests := []struct {
		name     string
		compiler *FilterCompiler
		filter   string
		wantSQL  string
		wantArgs []any
	}{
		{
			name: "nil mapping keeps field names", compiler: unmapped, filter: "total > 10",
			wantSQL: "SELECT id FROM orders o WHERE total > $1", wantArgs: []any{int64(10)},
		},
		{
			name: "nil mapping interpolates has", compiler: unmapped, filter: `tags:"blue"`,
			wantSQL: "SELECT id FROM orders o WHERE $1 = ANY(tags)", wantArgs: []any{"blue"},
		},
		{
			name: "nil mapping searches field names", compiler: unmapped, filter: `"blue"`,
			wantSQL: "SELECT id FROM orders o WHERE (name LIKE $1)", wantArgs: []any{"%blue%"},
		},
		{
			name: "partial mapping overrides one column", compiler: partial, filter: "total > 10",
			wantSQL: "SELECT id FROM orders o WHERE o.total > $1", wantArgs: []any{int64(10)},
		},
		{
			name: "partial mapping defaults the rest", compiler: partial, filter: "state = paid AND total > 10",
			wantSQL:  "SELECT id FROM orders o WHERE (state = $1 AND o.total > $2)",
			wantArgs: []any{"paid", int64(10)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			f, err := filter.Parse(test.filter)
			assert.NoError(t, err)
			gotSQL, gotArgs, gotErr := test.compiler.Where(psql.Select("id").From("orders o"), f).ToSql()
			assert.NoError(t, gotErr)
			assert.Equal(t, test.wantSQL, gotSQL)
			assert.Equal(t, test.wantArgs, gotArgs)
		})
	}
}

func TestNewFilterCompiler(t *testing.T) {
	t.Parallel()
	schema := testFilterSchema(t)
	tests := []struct {
		name    string
		schema  *filter.Schema
		fields  map[string]string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "full mapping",
			schema: schema,
			fields: map[string]string{
				"state":   "state",
				"total":   "total",
				"name":    "name",
				"deleted": "deleted",
				"tags":    "tags",
			},
			wantErr: assert.NoError,
		},
		{name: "partial mapping", schema: schema, fields: map[string]string{"state": "o.state"}, wantErr: assert.NoError},
		{name: "nil mapping", schema: schema, wantErr: assert.NoError},
		{name: "empty mapping", schema: schema, fields: map[string]string{}, wantErr: assert.NoError},
		{name: "uninitialized schema", fields: map[string]string{}, wantErr: assert.Error},
		{name: "uninitialized schema with nil mapping", wantErr: assert.Error},
		{name: "unknown mapping", schema: schema, fields: map[string]string{"other": "other"}, wantErr: assert.Error},
		{name: "invalid identifier", schema: schema, fields: map[string]string{"state": "state;drop"}, wantErr: assert.Error},
		{name: "empty identifier", schema: schema, fields: map[string]string{"state": ""}, wantErr: assert.Error},
		{name: "leading dot", schema: schema, fields: map[string]string{"state": ".state"}, wantErr: assert.Error},
		{name: "trailing dot", schema: schema, fields: map[string]string{"state": "state."}, wantErr: assert.Error},
		{name: "numeric start", schema: schema, fields: map[string]string{"state": "1state"}, wantErr: assert.Error},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewFilterCompiler(test.schema, test.fields)
			test.wantErr(t, err)
		})
	}
}
