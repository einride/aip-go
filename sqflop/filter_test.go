package sqflop

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/ysomad/flop"
	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/internal/assert"
)

func testFilterSchema(t *testing.T) *flop.FilterSchema {
	t.Helper()
	schema, err := flop.NewFilterSchema([]flop.FilterField{
		flop.NewFilterField("state", flop.Enum("pending", "paid"), false),
		flop.NewFilterField("total", flop.TypeInt, false),
		flop.NewFilterField("name", flop.TypeString, true),
		flop.NewFilterField("deleted", flop.TypeBool, false),
		flop.NewFilterField("tags", flop.TypeStringList, false),
	}, flop.FilterLimits{})
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

	// Every row selects from the same base, so only the filter varies.
	type args struct {
		filter string
	}
	tests := []struct {
		name string
		args args
		// call replaces the Where call for compiler defenses that the schema
		// rejects before compilation, so a filter can never reach them.
		call     func() (sq.Sqlizer, error)
		wantSQL  string
		wantArgs []any
		wantErr  assert.ErrorFunc
	}{
		{name: "empty", wantSQL: "SELECT id FROM orders o", wantErr: assert.NoError},
		{
			name: "has", args: args{filter: `tags:"blue"`},
			wantSQL:  "SELECT id FROM orders o WHERE $1 = ANY(o.tags)",
			wantArgs: []any{"blue"},
			wantErr:  assert.NoError,
		},
		{
			name: "negated has", args: args{filter: `NOT tags:"blue"`},
			wantSQL:  "SELECT id FROM orders o WHERE NOT ($1 = ANY(o.tags))",
			wantArgs: []any{"blue"},
			wantErr:  assert.NoError,
		},
		{
			name: "has inside a disjunction", args: args{filter: `tags:"blue" OR tags:"red"`},
			wantSQL:  "SELECT id FROM orders o WHERE ($1 = ANY(o.tags) OR $2 = ANY(o.tags))",
			wantArgs: []any{"blue", "red"},
			wantErr:  assert.NoError,
		},
		{
			name: "negation distributes over has", args: args{filter: `NOT (tags:"blue" OR total < 10)`},
			wantSQL:  "SELECT id FROM orders o WHERE (NOT ($1 = ANY(o.tags)) AND o.total >= $2)",
			wantArgs: []any{"blue", int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name: "comparisons", args: args{filter: "state = paid AND total >= 10"},
			wantSQL:  "SELECT id FROM orders o WHERE (o.state = $1 AND o.total >= $2)",
			wantArgs: []any{"paid", int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name:     "not equal",
			args:     args{filter: "total != 10"},
			wantSQL:  "SELECT id FROM orders o WHERE o.total <> $1",
			wantArgs: []any{int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name:     "less",
			args:     args{filter: "total < 10"},
			wantSQL:  "SELECT id FROM orders o WHERE o.total < $1",
			wantArgs: []any{int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name:     "less or equal",
			args:     args{filter: "total <= 10"},
			wantSQL:  "SELECT id FROM orders o WHERE o.total <= $1",
			wantArgs: []any{int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name:     "greater",
			args:     args{filter: "total > 10"},
			wantSQL:  "SELECT id FROM orders o WHERE o.total > $1",
			wantArgs: []any{int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name: "null", args: args{filter: "name != null"},
			wantSQL: "SELECT id FROM orders o WHERE o.name IS NOT NULL",
			wantErr: assert.NoError,
		},
		{
			name: "trailing wildcard", args: args{filter: `name = "book*"`},
			wantSQL: `SELECT id FROM orders o WHERE o.name LIKE $1 ESCAPE '\'`, wantArgs: []any{"book%"},
			wantErr: assert.NoError,
		},
		{
			name: "negative leading wildcard", args: args{filter: `name != "*book"`},
			wantSQL: `SELECT id FROM orders o WHERE o.name NOT LIKE $1 ESCAPE '\'`, wantArgs: []any{"%book"},
			wantErr: assert.NoError,
		},
		{
			// A wildcard is honoured wherever it sits, not only at a boundary.
			name: "interior wildcard", args: args{filter: `name = "a*b"`},
			wantSQL: `SELECT id FROM orders o WHERE o.name LIKE $1 ESCAPE '\'`, wantArgs: []any{"a%b"},
			wantErr: assert.NoError,
		},
		{
			// The reported filter: the underscore is data, not a LIKE
			// metacharacter, and must be escaped rather than rejected.
			name: "wildcard around an underscore", args: args{filter: `name = "*pay_10*"`},
			wantSQL: `SELECT id FROM orders o WHERE o.name LIKE $1 ESCAPE '\'`, wantArgs: []any{`%pay\_10%`},
			wantErr: assert.NoError,
		},
		{
			name: "escaped wildcard is a literal", args: args{filter: `name = "a\*b"`},
			wantSQL: "SELECT id FROM orders o WHERE o.name = $1", wantArgs: []any{"a*b"},
			wantErr: assert.NoError,
		},
		{
			name:    "metacharacters without a wildcard compare literally",
			args:    args{filter: `name = "100%_x"`},
			wantSQL: "SELECT id FROM orders o WHERE o.name = $1", wantArgs: []any{"100%_x"},
			wantErr: assert.NoError,
		},
		{
			name: "text sequence", args: args{filter: `"blue" "chair"`},
			wantSQL:  `SELECT id FROM orders o WHERE ((o.name LIKE $1 ESCAPE '\') AND (o.name LIKE $2 ESCAPE '\'))`,
			wantArgs: []any{"%blue%", "%chair%"},
			wantErr:  assert.NoError,
		},
		{
			name: "logical or", args: args{filter: `"blue" OR "chair"`},
			wantSQL:  `SELECT id FROM orders o WHERE ((o.name LIKE $1 ESCAPE '\') OR (o.name LIKE $2 ESCAPE '\'))`,
			wantArgs: []any{"%blue%", "%chair%"},
			wantErr:  assert.NoError,
		},
		{
			name: "negation", args: args{filter: "NOT (state = paid OR total < 10)"},
			wantSQL:  "SELECT id FROM orders o WHERE (o.state <> $1 AND o.total >= $2)",
			wantArgs: []any{"paid", int64(10)},
			wantErr:  assert.NoError,
		},
		{
			name: "negated text", args: args{filter: `NOT "blue"`},
			wantSQL: `SELECT id FROM orders o WHERE (o.name NOT LIKE $1 ESCAPE '\')`, wantArgs: []any{"%blue%"},
			wantErr: assert.NoError,
		},
		{
			// A search term carrying LIKE metacharacters matches them literally.
			name: "text search escapes metacharacters", args: args{filter: `"a%b"`},
			wantSQL: `SELECT id FROM orders o WHERE (o.name LIKE $1 ESCAPE '\')`, wantArgs: []any{`%a\%b%`},
			wantErr: assert.NoError,
		},
		{
			// An explicit wildcard replaces the implicit contains-anywhere match
			// rather than compounding with it.
			name: "text search with an explicit wildcard", args: args{filter: `"blue*"`},
			wantSQL: `SELECT id FROM orders o WHERE (o.name LIKE $1 ESCAPE '\')`, wantArgs: []any{"blue%"},
			wantErr: assert.NoError,
		},
		{name: "invalid filter", args: args{filter: "missing = 1"}, wantErr: assert.Error},
		{
			name: "unsupported left expression",
			call: func() (sq.Sqlizer, error) {
				return compiler.compile(expr.LogicalExpr{
					Left:  nil,
					Op:    expr.And,
					Right: expr.TextExpr{V: expr.StringValue("ok")},
				}, false)
			},
			wantErr: assert.Error,
		},
		{
			name: "unsupported right expression",
			call: func() (sq.Sqlizer, error) {
				return compiler.compile(expr.LogicalExpr{
					Left:  expr.TextExpr{V: expr.StringValue("ok")},
					Op:    expr.And,
					Right: nil,
				}, false)
			},
			wantErr: assert.Error,
		},
		{
			// The expression set is closed, so an unknown node can only be a nil one.
			name:    "unknown expression",
			call:    func() (sq.Sqlizer, error) { return compiler.compile(nil, false) },
			wantErr: assert.Error,
		},
		{
			name: "comparison on an unknown field",
			call: func() (sq.Sqlizer, error) {
				return compiler.comparison("missing", expr.Equal, expr.Int64Value(1))
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid comparison operator",
			call: func() (sq.Sqlizer, error) {
				return compiler.comparison("total", expr.Has, expr.Int64Value(1))
			},
			wantErr: assert.Error,
		},
		{
			name: "has on an unknown field",
			call: func() (sq.Sqlizer, error) {
				return compiler.has("missing", expr.StringValue("blue"), false)
			},
			wantErr: assert.Error,
		},
		{
			name: "has with a non-string value",
			call: func() (sq.Sqlizer, error) {
				return compiler.has("tags", expr.Int64Value(1), false)
			},
			wantErr: assert.Error,
		},
		{
			name: "text search with a non-string value",
			call: func() (sq.Sqlizer, error) {
				return compiler.text(expr.Int64Value(1), false)
			},
			wantErr: assert.Error,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if test.call != nil {
				_, gotErr := test.call()
				test.wantErr(t, gotErr)
				return
			}

			filter, err := flop.ParseFilter(test.args.filter)
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
	type args struct {
		compiler *FilterCompiler
		filter   string
	}
	tests := []struct {
		name     string
		args     args
		wantSQL  string
		wantArgs []any
	}{
		{
			name:    "nil mapping keeps field names",
			args:    args{compiler: unmapped, filter: "total > 10"},
			wantSQL: "SELECT id FROM orders o WHERE total > $1", wantArgs: []any{int64(10)},
		},
		{
			name:    "nil mapping interpolates has",
			args:    args{compiler: unmapped, filter: `tags:"blue"`},
			wantSQL: "SELECT id FROM orders o WHERE $1 = ANY(tags)", wantArgs: []any{"blue"},
		},
		{
			name:    "nil mapping searches field names",
			args:    args{compiler: unmapped, filter: `"blue"`},
			wantSQL: `SELECT id FROM orders o WHERE (name LIKE $1 ESCAPE '\')`, wantArgs: []any{"%blue%"},
		},
		{
			name:    "partial mapping overrides one column",
			args:    args{compiler: partial, filter: "total > 10"},
			wantSQL: "SELECT id FROM orders o WHERE o.total > $1", wantArgs: []any{int64(10)},
		},
		{
			name:     "partial mapping defaults the rest",
			args:     args{compiler: partial, filter: "state = paid AND total > 10"},
			wantSQL:  "SELECT id FROM orders o WHERE (state = $1 AND o.total > $2)",
			wantArgs: []any{"paid", int64(10)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			f, err := flop.ParseFilter(test.args.filter)
			assert.NoError(t, err)
			gotSQL, gotArgs, gotErr := test.args.compiler.Where(psql.Select("id").From("orders o"), f).ToSql()
			assert.NoError(t, gotErr)
			assert.Equal(t, test.wantSQL, gotSQL)
			assert.Equal(t, test.wantArgs, gotArgs)
		})
	}
}

func TestNewFilterCompiler(t *testing.T) {
	t.Parallel()
	schema := testFilterSchema(t)
	type args struct {
		schema *flop.FilterSchema
		fields map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorFunc
	}{
		{
			name: "full mapping",
			args: args{schema: schema, fields: map[string]string{
				"state":   "state",
				"total":   "total",
				"name":    "name",
				"deleted": "deleted",
				"tags":    "tags",
			}},
			wantErr: assert.NoError,
		},
		{name: "partial mapping", args: args{schema: schema, fields: map[string]string{"state": "o.state"}}, wantErr: assert.NoError},
		{name: "nil mapping", args: args{schema: schema}, wantErr: assert.NoError},
		{name: "empty mapping", args: args{schema: schema, fields: map[string]string{}}, wantErr: assert.NoError},
		{name: "uninitialized schema", args: args{fields: map[string]string{}}, wantErr: assert.Error},
		{name: "uninitialized schema with nil mapping", wantErr: assert.Error},
		{name: "unknown mapping", args: args{schema: schema, fields: map[string]string{"other": "other"}}, wantErr: assert.Error},
		{name: "invalid identifier", args: args{schema: schema, fields: map[string]string{"state": "state;drop"}}, wantErr: assert.Error},
		{name: "empty identifier", args: args{schema: schema, fields: map[string]string{"state": ""}}, wantErr: assert.Error},
		{name: "leading dot", args: args{schema: schema, fields: map[string]string{"state": ".state"}}, wantErr: assert.Error},
		{name: "trailing dot", args: args{schema: schema, fields: map[string]string{"state": "state."}}, wantErr: assert.Error},
		{name: "numeric start", args: args{schema: schema, fields: map[string]string{"state": "1state"}}, wantErr: assert.Error},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewFilterCompiler(test.args.schema, test.args.fields)
			test.wantErr(t, err)
		})
	}
}
