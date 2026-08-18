package flop_test

import (
	"math"
	"testing"
	"time"

	"github.com/ysomad/flop"
	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/internal/assert"
)

func TestNewFilterField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		got       flop.FilterField
		wantField flop.FilterField
		wantErr   assert.ErrorFunc
	}{
		{
			name:      "plain",
			got:       flop.NewFilterField("age", flop.TypeInt, false),
			wantField: flop.FilterField{Name: "age", Type: flop.TypeInt},
			wantErr:   assert.NoError,
		},
		{
			name:      "searchable",
			got:       flop.NewFilterField("name", flop.TypeString, true),
			wantField: flop.FilterField{Name: "name", Type: flop.TypeString, Search: true},
			wantErr:   assert.NoError,
		},
		{
			name:      "enum keeps its values",
			got:       flop.NewFilterField("state", flop.Enum("open", "closed"), false),
			wantField: flop.FilterField{Name: "state", Type: flop.Enum("open", "closed")},
			wantErr:   assert.NoError,
		},
		{
			// NewField declares, NewSchema validates.
			name:      "invalid declaration is not rejected",
			got:       flop.NewFilterField("bad-name", flop.FilterType{}, true),
			wantField: flop.FilterField{Name: "bad-name", Search: true},
			wantErr:   assert.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.wantField, test.got)
			_, err := flop.NewFilterSchema([]flop.FilterField{test.got}, flop.FilterLimits{})
			test.wantErr(t, err)
		})
	}
}

func TestNewFilterSchema(t *testing.T) {
	t.Parallel()

	type args struct {
		fields []flop.FilterField
		limits flop.FilterLimits
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "no fields"},
		{
			name: "invalid limits",
			args: args{
				fields: []flop.FilterField{flop.NewFilterField("id", flop.TypeInt, false)},
				limits: flop.FilterLimits{MaxNodes: -1},
			},
		},
		{
			name: "invalid field name",
			args: args{fields: []flop.FilterField{flop.NewFilterField("bad-name", flop.TypeInt, false)}},
		},
		{
			name: "duplicate field",
			args: args{fields: []flop.FilterField{
				flop.NewFilterField("id", flop.TypeInt, false),
				flop.NewFilterField("id", flop.TypeInt, false),
			}},
		},
		{
			name: "invalid field type",
			args: args{fields: []flop.FilterField{flop.NewFilterField("id", flop.FilterType{}, false)}},
		},
		{
			name: "empty enum",
			args: args{fields: []flop.FilterField{flop.NewFilterField("state", flop.Enum(), false)}},
		},
		{
			name: "empty enum value",
			args: args{fields: []flop.FilterField{flop.NewFilterField("state", flop.Enum(""), false)}},
		},
		{
			name: "duplicate enum value",
			args: args{fields: []flop.FilterField{flop.NewFilterField("state", flop.Enum("open", "open"), false)}},
		},
		{
			name: "non-string search field",
			args: args{fields: []flop.FilterField{flop.NewFilterField("id", flop.TypeInt, true)}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := flop.NewFilterSchema(test.args.fields, test.args.limits)
			assert.Error(t, err)
		})
	}
}

func TestFilterSchema_Validate(t *testing.T) {
	t.Parallel()

	schema, err := flop.NewFilterSchema([]flop.FilterField{
		flop.NewFilterField("name", flop.TypeString, true),
		flop.NewFilterField("active", flop.TypeBool, false),
		flop.NewFilterField("age", flop.TypeInt, false),
		flop.NewFilterField("count", flop.TypeUint, false),
		flop.NewFilterField("score", flop.TypeFloat, false),
		flop.NewFilterField("created_at", flop.TypeTime, false),
		flop.NewFilterField("wait", flop.TypeDuration, false),
		flop.NewFilterField("state", flop.Enum("open", "closed"), false),
		flop.NewFilterField("tags", flop.TypeStringList, false),
	}, flop.FilterLimits{MaxBytes: 128, MaxDepth: 2, MaxNodes: 3})
	assert.NoError(t, err)
	disabled := flop.DisabledFilter()
	nodeLimited, err := flop.NewFilterSchema(
		[]flop.FilterField{flop.NewFilterField("name", flop.TypeString, true)},
		flop.FilterLimits{MaxDepth: 10, MaxNodes: 2},
	)
	assert.NoError(t, err)
	withoutSearch, err := flop.NewFilterSchema(
		[]flop.FilterField{flop.NewFilterField("age", flop.TypeInt, false)},
		flop.FilterLimits{},
	)
	assert.NoError(t, err)

	type validated struct {
		canonical string
		// expr pins the tree where canonical cannot: the rendering is identical
		// across numeric kinds, so a skipped conversion would still match.
		expr expr.Expr
	}
	tests := []struct {
		name       string
		input      string
		filter     flop.Filter
		schema     *flop.FilterSchema
		wantFilter validated
		wantErr    assert.ErrorFunc
	}{
		{name: "empty", wantErr: assert.NoError},
		{
			name:       "has on a list",
			input:      `tags:"blue"`,
			wantFilter: validated{canonical: `tags : "blue"`},
			wantErr:    assert.NoError,
		},
		{
			name:       "has with a bare identifier",
			input:      "tags:blue",
			wantFilter: validated{canonical: `tags : "blue"`},
			wantErr:    assert.NoError,
		},
		{name: "has rejects a non-string", input: "tags:1", wantErr: assert.Error},
		{name: "has rejects null", input: "tags:null", wantErr: assert.Error},
		{name: "has is rejected on a string", input: `name:"x"`, wantErr: assert.Error},
		{name: "has is rejected on an int", input: "age:1", wantErr: assert.Error},
		{name: "has is rejected on an enum", input: "state:open", wantErr: assert.Error},
		{name: "list rejects equality", input: `tags = "blue"`, wantErr: assert.Error},
		{name: "list rejects ordering", input: `tags > "blue"`, wantErr: assert.Error},
		{
			name: "typed integer", input: `age = "42"`, wantFilter: validated{
				canonical: "age = 42",
				expr:      expr.ComparisonExpr{Field: "age", Op: expr.Equal, Value: expr.Int64Value(42)},
			},
			wantErr: assert.NoError,
		},
		{
			name: "typed unsigned integer", input: `count = "42"`, wantFilter: validated{
				canonical: "count = 42",
				expr:      expr.ComparisonExpr{Field: "count", Op: expr.Equal, Value: expr.Uint64Value(42)},
			},
			wantErr: assert.NoError,
		},
		{
			name: "typed float", input: `score = "1.5"`, wantFilter: validated{
				canonical: "score = 1.5",
				expr:      expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Float64Value(1.5)},
			},
			wantErr: assert.NoError,
		},
		{
			name:       "typed timestamp",
			input:      `created_at >= "2026-08-16T10:00:00+07:00"`,
			wantFilter: validated{canonical: "created_at >= 2026-08-16T03:00:00Z"},
			wantErr:    assert.NoError,
		},
		{
			name:       "typed duration",
			input:      `wait < "1.5s"`,
			wantFilter: validated{canonical: "wait < 1.5s"},
			wantErr:    assert.NoError,
		},
		{
			name: "native boolean",
			filter: flop.Filter{
				Expr: expr.ComparisonExpr{Field: "active", Op: expr.Equal, Value: expr.BoolValue(true)},
			},
			wantFilter: validated{canonical: "active = true"},
			wantErr:    assert.NoError,
		},
		{
			name: "native integer",
			filter: flop.Filter{
				Expr: expr.ComparisonExpr{Field: "age", Op: expr.Equal, Value: expr.Int64Value(42)},
			},
			wantFilter: validated{canonical: "age = 42"},
			wantErr:    assert.NoError,
		},
		{
			name: "unsigned converted to integer",
			filter: flop.Filter{
				Expr: expr.ComparisonExpr{Field: "age", Op: expr.Equal, Value: expr.Uint64Value(42)},
			},
			wantFilter: validated{
				canonical: "age = 42",
				expr:      expr.ComparisonExpr{Field: "age", Op: expr.Equal, Value: expr.Int64Value(42)},
			},
			wantErr: assert.NoError,
		},
		{
			name: "native unsigned integer",
			filter: flop.Filter{
				Expr: expr.ComparisonExpr{Field: "count", Op: expr.Equal, Value: expr.Uint64Value(42)},
			},
			wantFilter: validated{canonical: "count = 42"},
			wantErr:    assert.NoError,
		},
		{
			name: "integer converted to unsigned",
			filter: flop.Filter{
				Expr: expr.ComparisonExpr{Field: "count", Op: expr.Equal, Value: expr.Int64Value(42)},
			},
			wantFilter: validated{
				canonical: "count = 42",
				expr:      expr.ComparisonExpr{Field: "count", Op: expr.Equal, Value: expr.Uint64Value(42)},
			},
			wantErr: assert.NoError,
		},
		{
			name: "native float",
			filter: flop.Filter{
				Expr: expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Float64Value(1.5)},
			},
			wantFilter: validated{canonical: "score = 1.5"},
			wantErr:    assert.NoError,
		},
		{
			name: "integer converted to float",
			filter: flop.Filter{
				Expr: expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Int64Value(1)},
			},
			wantFilter: validated{
				canonical: "score = 1",
				expr:      expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Float64Value(1)},
			},
			wantErr: assert.NoError,
		},
		{
			name: "unsigned converted to float",
			filter: flop.Filter{
				Expr: expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Uint64Value(1)},
			},
			wantFilter: validated{
				canonical: "score = 1",
				expr:      expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Float64Value(1)},
			},
			wantErr: assert.NoError,
		},
		{
			name: "native timestamp",
			filter: flop.Filter{
				Expr: expr.ComparisonExpr{Field: "created_at", Op: expr.Equal, Value: expr.TimeValue(time.Unix(1, 0))},
			},
			wantFilter: validated{canonical: "created_at = 1970-01-01T00:00:01Z"},
			wantErr:    assert.NoError,
		},
		{
			name: "native duration",
			filter: flop.Filter{
				Expr: expr.ComparisonExpr{Field: "wait", Op: expr.Equal, Value: expr.DurationValue(time.Second)},
			},
			wantFilter: validated{canonical: "wait = 1s"},
			wantErr:    assert.NoError,
		},
		{
			name:       "null equality",
			filter:     flop.Filter{Expr: expr.ComparisonExpr{Field: "age", Op: expr.Equal, Value: expr.NullValue()}},
			wantFilter: validated{canonical: "age = null"},
			wantErr:    assert.NoError,
		},
		{name: "text search", input: "42", wantFilter: validated{canonical: `"42"`}, wantErr: assert.NoError},
		{
			name:       "negation",
			input:      `NOT name = "x"`,
			wantFilter: validated{canonical: `NOT (name = "x")`},
			wantErr:    assert.NoError,
		},
		{
			name:       "enum",
			input:      "state = open",
			wantFilter: validated{canonical: `state = "open"`},
			wantErr:    assert.NoError,
		},
		{
			name:    "zero-value schema",
			schema:  new(flop.FilterSchema),
			filter:  flop.Filter{Expr: expr.TextExpr{V: expr.StringValue("x")}},
			wantErr: assert.Error,
		},
		{
			name:    "disabled schema",
			schema:  disabled,
			filter:  flop.Filter{Expr: expr.TextExpr{V: expr.StringValue("x")}},
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
			filter: flop.Filter{
				Expr: expr.ComparisonExpr{Field: "score", Op: expr.Equal, Value: expr.Float64Value(math.NaN())},
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid native timestamp",
			filter: flop.Filter{
				Expr: expr.ComparisonExpr{Field: "created_at", Op: expr.Equal, Value: expr.TimeValue(time.Time{})},
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid comparison operator",
			filter: flop.Filter{
				Expr: expr.ComparisonExpr{Field: "age", Op: expr.ComparisonOp(99), Value: expr.Int64Value(1)},
			},
			wantErr: assert.Error,
		},
		{
			name:    "invalid null operator",
			filter:  flop.Filter{Expr: expr.ComparisonExpr{Field: "age", Op: expr.Greater, Value: expr.NullValue()}},
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
			filter: flop.Filter{Expr: expr.LogicalExpr{
				Left: expr.TextExpr{
					V: expr.StringValue("a"),
				}, Op: expr.LogicalOp(99), Right: expr.TextExpr{V: expr.StringValue("b")},
			}},
			wantErr: assert.Error,
		},
		{
			name: "invalid left expression",
			filter: flop.Filter{Expr: expr.LogicalExpr{
				Left: nil, Op: expr.And, Right: expr.TextExpr{V: expr.StringValue("b")},
			}},
			wantErr: assert.Error,
		},
		{
			name: "invalid right expression",
			filter: flop.Filter{Expr: expr.LogicalExpr{
				Left: expr.TextExpr{V: expr.StringValue("a")}, Op: expr.And, Right: nil,
			}},
			wantErr: assert.Error,
		},
		{name: "invalid negated expression", filter: flop.Filter{Expr: expr.NotExpr{}}, wantErr: assert.Error},
		{
			name:    "invalid text value",
			filter:  flop.Filter{Expr: expr.TextExpr{V: expr.NullValue()}},
			wantErr: assert.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			f := test.filter
			if test.input != "" {
				var err error
				f, err = flop.ParseFilter(test.input)
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
			assert.Equal(t, test.wantFilter.canonical, got.String())
			if test.wantFilter.expr != nil {
				assert.Equal(t, test.wantFilter.expr, got.Expr)
			}
		})
	}
}

func TestFilterSchema_ValidateNil(t *testing.T) {
	t.Parallel()
	var schema *flop.FilterSchema
	assert.False(t, schema.Initialized())
	_, err := schema.Validate(flop.Filter{Expr: expr.TextExpr{V: expr.StringValue("x")}})
	assert.Error(t, err)
}

func TestFilterSchema_Fields(t *testing.T) {
	t.Parallel()
	schema, err := inspectionSchema()
	assert.NoError(t, err)

	tests := []struct {
		name   string
		schema *flop.FilterSchema
		want   []flop.FilterField
	}{
		{
			name:   "declared",
			schema: schema,
			want: []flop.FilterField{
				flop.NewFilterField("name", flop.TypeString, true),
				flop.NewFilterField("age", flop.TypeInt, false),
			},
		},
		{name: "nil schema", schema: nil, want: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.schema.Fields())
		})
	}

	fields := schema.Fields()
	fields[0].Name = "changed"
	assert.Equal(t, "name", schema.Fields()[0].Name)
}

func TestFilterSchema_SearchFields(t *testing.T) {
	t.Parallel()
	schema, err := inspectionSchema()
	assert.NoError(t, err)

	tests := []struct {
		name   string
		schema *flop.FilterSchema
		want   []string
	}{
		{name: "declared", schema: schema, want: []string{"name"}},
		{name: "nil schema", schema: nil, want: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.schema.SearchFields())
		})
	}

	searchFields := schema.SearchFields()
	searchFields[0] = "changed"
	assert.Equal(t, []string{"name"}, schema.SearchFields())
}

func TestFilterSchema_IsString(t *testing.T) {
	t.Parallel()
	schema, err := inspectionSchema()
	assert.NoError(t, err)
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		schema *flop.FilterSchema
		args   args
		want   bool
	}{
		{name: "string", schema: schema, args: args{name: "name"}, want: true},
		{name: "non-string", schema: schema, args: args{name: "age"}},
		{name: "missing", schema: schema, args: args{name: "missing"}},
		{name: "nil schema", schema: nil, args: args{name: "name"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.schema.IsString(test.args.name))
		})
	}
}

func inspectionSchema() (*flop.FilterSchema, error) {
	return flop.NewFilterSchema([]flop.FilterField{
		flop.NewFilterField("name", flop.TypeString, true),
		flop.NewFilterField("age", flop.TypeInt, false),
	}, flop.FilterLimits{})
}
