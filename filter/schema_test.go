package filter_test

import (
	"math"
	"testing"
	"time"

	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/filter"
	"github.com/ysomad/flop/internal/assert"
)

func TestNewFilterField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		got     filter.Field
		want    filter.Field
		wantErr assert.ErrorFunc
	}{
		{
			name:    "plain",
			got:     filter.NewField("age", filter.TypeInt, false),
			want:    filter.Field{Name: "age", Type: filter.TypeInt},
			wantErr: assert.NoError,
		},
		{
			name:    "searchable",
			got:     filter.NewField("name", filter.TypeString, true),
			want:    filter.Field{Name: "name", Type: filter.TypeString, Search: true},
			wantErr: assert.NoError,
		},
		{
			name:    "enum keeps its values",
			got:     filter.NewField("state", filter.Enum("open", "closed"), false),
			want:    filter.Field{Name: "state", Type: filter.Enum("open", "closed")},
			wantErr: assert.NoError,
		},
		{
			// NewField declares, NewSchema validates.
			name:    "invalid declaration is not rejected",
			got:     filter.NewField("bad-name", filter.Type{}, true),
			want:    filter.Field{Name: "bad-name", Search: true},
			wantErr: assert.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.got)
			_, err := filter.NewSchema([]filter.Field{test.got}, filter.Limits{})
			test.wantErr(t, err)
		})
	}
}

func TestNewFilterSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		fields []filter.Field
		limits filter.Limits
	}{
		{name: "no fields"},
		{
			name:   "invalid limits",
			fields: []filter.Field{filter.NewField("id", filter.TypeInt, false)},
			limits: filter.Limits{MaxNodes: -1},
		},
		{name: "invalid field name", fields: []filter.Field{filter.NewField("bad-name", filter.TypeInt, false)}},
		{
			name: "duplicate field",
			fields: []filter.Field{
				filter.NewField("id", filter.TypeInt, false),
				filter.NewField("id", filter.TypeInt, false),
			},
		},
		{name: "invalid field type", fields: []filter.Field{filter.NewField("id", filter.Type{}, false)}},
		{name: "empty enum", fields: []filter.Field{filter.NewField("state", filter.Enum(), false)}},
		{name: "empty enum value", fields: []filter.Field{filter.NewField("state", filter.Enum(""), false)}},
		{
			name:   "duplicate enum value",
			fields: []filter.Field{filter.NewField("state", filter.Enum("open", "open"), false)},
		},
		{name: "non-string search field", fields: []filter.Field{filter.NewField("id", filter.TypeInt, true)}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := filter.NewSchema(test.fields, test.limits)
			assert.Error(t, err)
		})
	}
}

func TestFilterSchema_Validate(t *testing.T) {
	t.Parallel()

	schema, err := filter.NewSchema([]filter.Field{
		filter.NewField("name", filter.TypeString, true),
		filter.NewField("active", filter.TypeBool, false),
		filter.NewField("age", filter.TypeInt, false),
		filter.NewField("count", filter.TypeUint, false),
		filter.NewField("score", filter.TypeFloat, false),
		filter.NewField("created_at", filter.TypeTime, false),
		filter.NewField("wait", filter.TypeDuration, false),
		filter.NewField("state", filter.Enum("open", "closed"), false),
		filter.NewField("tags", filter.TypeStringList, false),
	}, filter.Limits{MaxBytes: 128, MaxDepth: 2, MaxNodes: 3})
	assert.NoError(t, err)
	disabled := filter.Disabled()
	nodeLimited, err := filter.NewSchema(
		[]filter.Field{filter.NewField("name", filter.TypeString, true)},
		filter.Limits{MaxDepth: 10, MaxNodes: 2},
	)
	assert.NoError(t, err)
	withoutSearch, err := filter.NewSchema(
		[]filter.Field{filter.NewField("age", filter.TypeInt, false)},
		filter.Limits{},
	)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		input   string
		filter  filter.Filter
		schema  *filter.Schema
		want    string
		wantErr assert.ErrorFunc
	}{
		{name: "empty", wantErr: assert.NoError},
		{name: "has on a list", input: `tags:"blue"`, want: `tags : "blue"`, wantErr: assert.NoError},
		{name: "has with a bare identifier", input: "tags:blue", want: `tags : "blue"`, wantErr: assert.NoError},
		{name: "has rejects a non-string", input: "tags:1", wantErr: assert.Error},
		{name: "has rejects null", input: "tags:null", wantErr: assert.Error},
		{name: "has is rejected on a string", input: `name:"x"`, wantErr: assert.Error},
		{name: "has is rejected on an int", input: "age:1", wantErr: assert.Error},
		{name: "has is rejected on an enum", input: "state:open", wantErr: assert.Error},
		{name: "list rejects equality", input: `tags = "blue"`, wantErr: assert.Error},
		{name: "list rejects ordering", input: `tags > "blue"`, wantErr: assert.Error},
		{name: "typed integer", input: `age = "42"`, want: "age = 42", wantErr: assert.NoError},
		{name: "typed unsigned integer", input: `count = "42"`, want: "count = 42", wantErr: assert.NoError},
		{name: "typed float", input: `score = "1.5"`, want: "score = 1.5", wantErr: assert.NoError},
		{
			name:    "typed timestamp",
			input:   `created_at >= "2026-08-16T10:00:00+07:00"`,
			want:    "created_at >= 2026-08-16T03:00:00Z",
			wantErr: assert.NoError,
		},
		{name: "typed duration", input: `wait < "1.5s"`, want: "wait < 1.5s", wantErr: assert.NoError},
		{
			name: "native boolean",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "active", Op: expr.Equal, Value: expr.BoolValue(true)},
			},
			want:    "active = true",
			wantErr: assert.NoError,
		},
		{
			name:    "native integer",
			filter:  filter.Filter{Expr: expr.ComparisonExpr{Field: "age", Op: expr.Equal, Value: expr.Int64Value(42)}},
			want:    "age = 42",
			wantErr: assert.NoError,
		},
		{
			name: "unsigned converted to integer",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "age", Op: expr.Equal, Value: expr.Uint64Value(42)},
			},
			want:    "age = 42",
			wantErr: assert.NoError,
		},
		{
			name: "native unsigned integer",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "count", Op: expr.Equal, Value: expr.Uint64Value(42)},
			},
			want:    "count = 42",
			wantErr: assert.NoError,
		},
		{
			name: "integer converted to unsigned",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "count", Op: expr.Equal, Value: expr.Int64Value(42)},
			},
			want:    "count = 42",
			wantErr: assert.NoError,
		},
		{
			name: "native float",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Float64Value(1.5)},
			},
			want:    "score = 1.5",
			wantErr: assert.NoError,
		},
		{
			name: "integer converted to float",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Int64Value(1)},
			},
			want:    "score = 1",
			wantErr: assert.NoError,
		},
		{
			name: "unsigned converted to float",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Uint64Value(1)},
			},
			want:    "score = 1",
			wantErr: assert.NoError,
		},
		{
			name: "native timestamp",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "created_at", Op: expr.Equal, Value: expr.TimeValue(time.Unix(1, 0))},
			},
			want:    "created_at = 1970-01-01T00:00:01Z",
			wantErr: assert.NoError,
		},
		{
			name: "native duration",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "wait", Op: expr.Equal, Value: expr.DurationValue(time.Second)},
			},
			want:    "wait = 1s",
			wantErr: assert.NoError,
		},
		{
			name:    "null equality",
			filter:  filter.Filter{Expr: expr.ComparisonExpr{Field: "age", Op: expr.Equal, Value: expr.NullValue()}},
			want:    "age = null",
			wantErr: assert.NoError,
		},
		{name: "text search", input: "42", want: `"42"`, wantErr: assert.NoError},
		{name: "negation", input: `NOT name = "x"`, want: `NOT (name = "x")`, wantErr: assert.NoError},
		{name: "enum", input: "state = open", want: `state = "open"`, wantErr: assert.NoError},
		{
			name:    "zero-value schema",
			schema:  new(filter.Schema),
			filter:  filter.Filter{Expr: expr.TextExpr{V: expr.StringValue("x")}},
			wantErr: assert.Error,
		},
		{
			name:    "disabled schema",
			schema:  disabled,
			filter:  filter.Filter{Expr: expr.TextExpr{V: expr.StringValue("x")}},
			wantErr: assert.Error,
		},
		{name: "unknown field", input: "missing = 1", wantErr: assert.Error},
		{name: "invalid enum", input: "state = missing", wantErr: assert.Error},
		{name: "invalid enum operator", input: "state > open", wantErr: assert.Error},
		{name: "invalid boolean operator", input: "active > true", wantErr: assert.Error},
		{name: "invalid unsigned integer", input: `count = "-1"`, wantErr: assert.Error},
		{name: "invalid float", input: `score = "NaN"`, wantErr: assert.Error},
		{name: "invalid timestamp", input: `created_at = "tomorrow"`, wantErr: assert.Error},
		{name: "invalid duration", input: `wait = "soon"`, wantErr: assert.Error},
		{
			name: "invalid native float",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Float64Value(math.NaN())},
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid native timestamp",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "created_at", Op: expr.Equal, Value: expr.TimeValue(time.Time{})},
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid comparison operator",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "age", Op: expr.ComparisonOp(99), Value: expr.Int64Value(1)},
			},
			wantErr: assert.Error,
		},
		{
			name:    "invalid null operator",
			filter:  filter.Filter{Expr: expr.ComparisonExpr{Field: "age", Op: expr.Greater, Value: expr.NullValue()}},
			wantErr: assert.Error,
		},
		{
			name:    "too large",
			input:   `name = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`,
			wantErr: assert.Error,
		},
		{name: "too deep", input: `NOT (NOT (name = "x"))`, wantErr: assert.Error},
		{name: "too many expressions", schema: nodeLimited, input: `name = "a" AND name = "b"`, wantErr: assert.Error},
		{name: "text search disabled", schema: withoutSearch, input: `"x"`, wantErr: assert.Error},
		{
			name: "invalid logical operator",
			filter: filter.Filter{Expr: expr.LogicalExpr{
				Left: expr.TextExpr{
					V: expr.StringValue("a"),
				}, Op: expr.LogicalOp(99), Right: expr.TextExpr{V: expr.StringValue("b")},
			}},
			wantErr: assert.Error,
		},
		{
			name: "invalid left expression",
			filter: filter.Filter{Expr: expr.LogicalExpr{
				Left: nil, Op: expr.And, Right: expr.TextExpr{V: expr.StringValue("b")},
			}},
			wantErr: assert.Error,
		},
		{
			name: "invalid right expression",
			filter: filter.Filter{Expr: expr.LogicalExpr{
				Left: expr.TextExpr{V: expr.StringValue("a")}, Op: expr.And, Right: nil,
			}},
			wantErr: assert.Error,
		},
		{name: "invalid negated expression", filter: filter.Filter{Expr: expr.NotExpr{}}, wantErr: assert.Error},
		{
			name:    "invalid text value",
			filter:  filter.Filter{Expr: expr.TextExpr{V: expr.NullValue()}},
			wantErr: assert.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			f := test.filter
			if test.input != "" {
				var err error
				f, err = filter.Parse(test.input)
				assert.NoError(t, err)
			}
			validator := schema
			if test.schema != nil {
				validator = test.schema
			}
			got, err := validator.Validate(f)
			test.wantErr(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, test.want, got.String())
		})
	}
}

func TestFilterSchema_ValidateNil(t *testing.T) {
	t.Parallel()
	var schema *filter.Schema
	assert.False(t, schema.Initialized())
	_, err := schema.Validate(filter.Filter{Expr: expr.TextExpr{V: expr.StringValue("x")}})
	assert.Error(t, err)
}

func TestFilterSchema_Fields(t *testing.T) {
	t.Parallel()
	schema, err := inspectionSchema()
	assert.NoError(t, err)

	fields := schema.Fields()
	fields[0].Name = "changed"
	assert.Equal(t, "name", schema.Fields()[0].Name)
}

func TestFilterSchema_SearchFields(t *testing.T) {
	t.Parallel()
	schema, err := inspectionSchema()
	assert.NoError(t, err)
	searchFields := schema.SearchFields()
	searchFields[0] = "changed"
	assert.Equal(t, []string{"name"}, schema.SearchFields())
}

func TestFilterSchema_IsString(t *testing.T) {
	t.Parallel()
	schema, err := inspectionSchema()
	assert.NoError(t, err)
	tests := []struct {
		name  string
		field string
		want  bool
	}{
		{name: "string", field: "name", want: true},
		{name: "non-string", field: "age"},
		{name: "missing", field: "missing"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, schema.IsString(test.field))
		})
	}
}

func inspectionSchema() (*filter.Schema, error) {
	return filter.NewSchema([]filter.Field{
		filter.NewField("name", filter.TypeString, true),
		filter.NewField("age", filter.TypeInt, false),
	}, filter.Limits{})
}
