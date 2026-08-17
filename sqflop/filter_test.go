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

func TestFilterCompiler_Apply(t *testing.T) {
	t.Parallel()
	compiler, err := NewFilterCompiler(testFilterSchema(t), map[string]string{
		"state": "o.state", "total": "o.total", "name": "o.name", "deleted": "o.deleted",
		"tags": "o.tags",
	})
	assert.NoError(t, err)
	tests := []struct {
		name     string
		filter   string
		wantSQL  string
		wantArgs []any
		wantErr  bool
	}{
		{name: "empty", wantSQL: "SELECT id FROM orders o"},
		{
			name: "has", filter: `tags:"blue"`,
			wantSQL:  "SELECT id FROM orders o WHERE $1 = ANY(o.tags)",
			wantArgs: []any{"blue"},
		},
		{
			name: "negated has", filter: `NOT tags:"blue"`,
			wantSQL:  "SELECT id FROM orders o WHERE NOT ($1 = ANY(o.tags))",
			wantArgs: []any{"blue"},
		},
		{
			name: "has inside a disjunction", filter: `tags:"blue" OR tags:"red"`,
			wantSQL:  "SELECT id FROM orders o WHERE ($1 = ANY(o.tags) OR $2 = ANY(o.tags))",
			wantArgs: []any{"blue", "red"},
		},
		{
			name: "negation distributes over has", filter: `NOT (tags:"blue" OR total < 10)`,
			wantSQL:  "SELECT id FROM orders o WHERE (NOT ($1 = ANY(o.tags)) AND o.total >= $2)",
			wantArgs: []any{"blue", int64(10)},
		},
		{
			name: "comparisons", filter: "state = paid AND total >= 10",
			wantSQL:  "SELECT id FROM orders o WHERE (o.state = $1 AND o.total >= $2)",
			wantArgs: []any{"paid", int64(10)},
		},
		{
			name:     "not equal",
			filter:   "total != 10",
			wantSQL:  "SELECT id FROM orders o WHERE o.total <> $1",
			wantArgs: []any{int64(10)},
		},
		{
			name:     "less",
			filter:   "total < 10",
			wantSQL:  "SELECT id FROM orders o WHERE o.total < $1",
			wantArgs: []any{int64(10)},
		},
		{
			name:     "less or equal",
			filter:   "total <= 10",
			wantSQL:  "SELECT id FROM orders o WHERE o.total <= $1",
			wantArgs: []any{int64(10)},
		},
		{
			name:     "greater",
			filter:   "total > 10",
			wantSQL:  "SELECT id FROM orders o WHERE o.total > $1",
			wantArgs: []any{int64(10)},
		},
		{name: "null", filter: "name != null", wantSQL: "SELECT id FROM orders o WHERE o.name IS NOT NULL"},
		{
			name: "boundary wildcard", filter: `name = "book*"`,
			wantSQL: "SELECT id FROM orders o WHERE o.name LIKE $1", wantArgs: []any{"book%"},
		},
		{
			name: "negative boundary wildcard", filter: `name != "*book"`,
			wantSQL: "SELECT id FROM orders o WHERE o.name NOT LIKE $1", wantArgs: []any{"%book"},
		},
		{
			name: "text sequence", filter: `"blue" "chair"`,
			wantSQL:  "SELECT id FROM orders o WHERE ((o.name LIKE $1) AND (o.name LIKE $2))",
			wantArgs: []any{"%blue%", "%chair%"},
		},
		{
			name: "logical or", filter: `"blue" OR "chair"`,
			wantSQL:  "SELECT id FROM orders o WHERE ((o.name LIKE $1) OR (o.name LIKE $2))",
			wantArgs: []any{"%blue%", "%chair%"},
		},
		{
			name: "negation", filter: "NOT (state = paid OR total < 10)",
			wantSQL:  "SELECT id FROM orders o WHERE (o.state <> $1 AND o.total >= $2)",
			wantArgs: []any{"paid", int64(10)},
		},
		{
			name: "negated text", filter: `NOT "blue"`,
			wantSQL: "SELECT id FROM orders o WHERE (o.name NOT LIKE $1)", wantArgs: []any{"%blue%"},
		},
		{name: "invalid filter", filter: "missing = 1", wantErr: true},
		{name: "unsupported pattern", filter: `name = "*a_b*"`, wantErr: true},
		{name: "unsupported text pattern", filter: `"a%b"`, wantErr: true},
		{name: "unsupported left expression", filter: `"a%b" AND "ok"`, wantErr: true},
		{name: "unsupported right expression", filter: `"ok" AND "a%b"`, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			filter, err := filter.Parse(test.filter)
			assert.NoError(t, err)
			gotSQL, gotArgs, gotErr := compiler.Apply(psql.Select("id").From("orders o"), filter).ToSql()
			if test.wantErr {
				assert.Error(t, gotErr)
				return
			}
			assert.NoError(t, gotErr)
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
			gotSQL, gotArgs, gotErr := test.compiler.Apply(psql.Select("id").From("orders o"), f).ToSql()
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
		wantErr bool
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
		},
		{name: "partial mapping", schema: schema, fields: map[string]string{"state": "o.state"}},
		{name: "nil mapping", schema: schema},
		{name: "empty mapping", schema: schema, fields: map[string]string{}},
		{name: "uninitialized schema", fields: map[string]string{}, wantErr: true},
		{name: "uninitialized schema with nil mapping", wantErr: true},
		{name: "unknown mapping", schema: schema, fields: map[string]string{"other": "other"}, wantErr: true},
		{name: "invalid identifier", schema: schema, fields: map[string]string{"state": "state;drop"}, wantErr: true},
		{name: "empty identifier", schema: schema, fields: map[string]string{"state": ""}, wantErr: true},
		{name: "leading dot", schema: schema, fields: map[string]string{"state": ".state"}, wantErr: true},
		{name: "trailing dot", schema: schema, fields: map[string]string{"state": "state."}, wantErr: true},
		{name: "numeric start", schema: schema, fields: map[string]string{"state": "1state"}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewFilterCompiler(test.schema, test.fields)
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
