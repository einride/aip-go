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
		name   string
		got    filter.Field
		want   filter.Field
		schema bool
	}{
		{
			name:   "plain",
			got:    filter.NewField("age", filter.TypeInt, false),
			want:   filter.Field{Name: "age", Type: filter.TypeInt},
			schema: true,
		},
		{
			name:   "searchable",
			got:    filter.NewField("name", filter.TypeString, true),
			want:   filter.Field{Name: "name", Type: filter.TypeString, Search: true},
			schema: true,
		},
		{
			name:   "enum keeps its values",
			got:    filter.NewField("state", filter.Enum("open", "closed"), false),
			want:   filter.Field{Name: "state", Type: filter.Enum("open", "closed")},
			schema: true,
		},
		{
			// NewField declares, NewSchema validates.
			name: "invalid declaration is not rejected",
			got:  filter.NewField("bad-name", filter.Type{}, true),
			want: filter.Field{Name: "bad-name", Search: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.got)
			_, err := filter.NewSchema([]filter.Field{test.got}, filter.Limits{})
			if test.schema {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
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
		{name: "duplicate enum value", fields: []filter.Field{filter.NewField("state", filter.Enum("open", "open"), false)}},
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
		wantErr bool
	}{
		{name: "empty"},
		{name: "has on a list", input: `tags:"blue"`, want: `tags : "blue"`},
		{name: "has with a bare identifier", input: "tags:blue", want: `tags : "blue"`},
		{name: "has rejects a non-string", input: "tags:1", wantErr: true},
		{name: "has rejects null", input: "tags:null", wantErr: true},
		{name: "has is rejected on a string", input: `name:"x"`, wantErr: true},
		{name: "has is rejected on an int", input: "age:1", wantErr: true},
		{name: "has is rejected on an enum", input: "state:open", wantErr: true},
		{name: "list rejects equality", input: `tags = "blue"`, wantErr: true},
		{name: "list rejects ordering", input: `tags > "blue"`, wantErr: true},
		{name: "typed integer", input: `age = "42"`, want: "age = 42"},
		{name: "typed unsigned integer", input: `count = "42"`, want: "count = 42"},
		{name: "typed float", input: `score = "1.5"`, want: "score = 1.5"},
		{
			name:  "typed timestamp",
			input: `created_at >= "2026-08-16T10:00:00+07:00"`,
			want:  "created_at >= 2026-08-16T03:00:00Z",
		},
		{name: "typed duration", input: `wait < "1.5s"`, want: "wait < 1.5s"},
		{
			name: "native boolean",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "active", Op: expr.Equal, Value: expr.BoolValue(true)},
			},
			want: "active = true",
		},
		{
			name:   "native integer",
			filter: filter.Filter{Expr: expr.ComparisonExpr{Field: "age", Op: expr.Equal, Value: expr.Int64Value(42)}},
			want:   "age = 42",
		},
		{
			name:   "unsigned converted to integer",
			filter: filter.Filter{Expr: expr.ComparisonExpr{Field: "age", Op: expr.Equal, Value: expr.Uint64Value(42)}},
			want:   "age = 42",
		},
		{
			name: "native unsigned integer",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "count", Op: expr.Equal, Value: expr.Uint64Value(42)},
			},
			want: "count = 42",
		},
		{
			name: "integer converted to unsigned",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "count", Op: expr.Equal, Value: expr.Int64Value(42)},
			},
			want: "count = 42",
		},
		{
			name: "native float",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Float64Value(1.5)},
			},
			want: "score = 1.5",
		},
		{
			name:   "integer converted to float",
			filter: filter.Filter{Expr: expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Int64Value(1)}},
			want:   "score = 1",
		},
		{
			name: "unsigned converted to float",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Uint64Value(1)},
			},
			want: "score = 1",
		},
		{
			name: "native timestamp",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "created_at", Op: expr.Equal, Value: expr.TimeValue(time.Unix(1, 0))},
			},
			want: "created_at = 1970-01-01T00:00:01Z",
		},
		{
			name: "native duration",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "wait", Op: expr.Equal, Value: expr.DurationValue(time.Second)},
			},
			want: "wait = 1s",
		},
		{
			name:   "null equality",
			filter: filter.Filter{Expr: expr.ComparisonExpr{Field: "age", Op: expr.Equal, Value: expr.NullValue()}},
			want:   "age = null",
		},
		{name: "text search", input: "42", want: `"42"`},
		{name: "negation", input: `NOT name = "x"`, want: `NOT (name = "x")`},
		{name: "enum", input: "state = open", want: `state = "open"`},
		{
			name:    "zero-value schema",
			schema:  new(filter.Schema),
			filter:  filter.Filter{Expr: expr.TextExpr{V: expr.StringValue("x")}},
			wantErr: true,
		},
		{
			name:    "disabled schema",
			schema:  disabled,
			filter:  filter.Filter{Expr: expr.TextExpr{V: expr.StringValue("x")}},
			wantErr: true,
		},
		{name: "unknown field", input: "missing = 1", wantErr: true},
		{name: "invalid enum", input: "state = missing", wantErr: true},
		{name: "invalid enum operator", input: "state > open", wantErr: true},
		{name: "invalid boolean operator", input: "active > true", wantErr: true},
		{name: "invalid unsigned integer", input: `count = "-1"`, wantErr: true},
		{name: "invalid float", input: `score = "NaN"`, wantErr: true},
		{name: "invalid timestamp", input: `created_at = "tomorrow"`, wantErr: true},
		{name: "invalid duration", input: `wait = "soon"`, wantErr: true},
		{
			name: "invalid native float",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Float64Value(math.NaN())},
			},
			wantErr: true,
		},
		{
			name: "invalid native timestamp",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "created_at", Op: expr.Equal, Value: expr.TimeValue(time.Time{})},
			},
			wantErr: true,
		},
		{
			name: "invalid comparison operator",
			filter: filter.Filter{
				Expr: expr.ComparisonExpr{Field: "age", Op: expr.ComparisonOp(99), Value: expr.Int64Value(1)},
			},
			wantErr: true,
		},
		{
			name:    "invalid null operator",
			filter:  filter.Filter{Expr: expr.ComparisonExpr{Field: "age", Op: expr.Greater, Value: expr.NullValue()}},
			wantErr: true,
		},
		{
			name:    "too large",
			input:   `name = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`,
			wantErr: true,
		},
		{name: "too deep", input: `NOT (NOT (name = "x"))`, wantErr: true},
		{name: "too many expressions", schema: nodeLimited, input: `name = "a" AND name = "b"`, wantErr: true},
		{name: "text search disabled", schema: withoutSearch, input: `"x"`, wantErr: true},
		{
			name: "invalid logical operator",
			filter: filter.Filter{Expr: expr.LogicalExpr{
				Left: expr.TextExpr{
					V: expr.StringValue("a"),
				}, Op: expr.LogicalOp(99), Right: expr.TextExpr{V: expr.StringValue("b")},
			}},
			wantErr: true,
		},
		{
			name: "invalid left expression",
			filter: filter.Filter{Expr: expr.LogicalExpr{
				Left: nil, Op: expr.And, Right: expr.TextExpr{V: expr.StringValue("b")},
			}},
			wantErr: true,
		},
		{
			name: "invalid right expression",
			filter: filter.Filter{Expr: expr.LogicalExpr{
				Left: expr.TextExpr{V: expr.StringValue("a")}, Op: expr.And, Right: nil,
			}},
			wantErr: true,
		},
		{name: "invalid negated expression", filter: filter.Filter{Expr: expr.NotExpr{}}, wantErr: true},
		{
			name:    "invalid text value",
			filter:  filter.Filter{Expr: expr.TextExpr{V: expr.NullValue()}},
			wantErr: true,
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
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
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
