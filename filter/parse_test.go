package filter_test

import (
	"math"
	"strings"
	"testing"

	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/filter"
	"github.com/ysomad/flop/internal/assert"
)

func TestParseFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    filter.Filter
		wantErr bool
	}{
		{name: "empty", input: " \n\t", want: filter.Filter{}},
		{
			name:  "comparison",
			input: "price >= 12.5",
			want: filter.Filter{Expr: expr.ComparisonExpr{
				Field: "price", Op: expr.GreaterOrEqual, Value: expr.Float64Value(12.5),
			}},
		},
		{
			name:  "all comparison operators",
			input: "a != 1 AND b < 2 AND c <= 3 AND d > 4",
			want: filter.Filter{Expr: expr.LogicalExpr{
				Left: expr.LogicalExpr{
					Left: expr.LogicalExpr{
						Left:  expr.ComparisonExpr{Field: "a", Op: expr.NotEqual, Value: expr.Int64Value(1)},
						Op:    expr.And,
						Right: expr.ComparisonExpr{Field: "b", Op: expr.Less, Value: expr.Int64Value(2)},
					},
					Op:    expr.And,
					Right: expr.ComparisonExpr{Field: "c", Op: expr.LessOrEqual, Value: expr.Int64Value(3)},
				},
				Op:    expr.And,
				Right: expr.ComparisonExpr{Field: "d", Op: expr.Greater, Value: expr.Int64Value(4)},
			}},
		},
		{
			name:  "minimum int64",
			input: "value = -9223372036854775808",
			want: filter.Filter{Expr: expr.ComparisonExpr{
				Field: "value", Op: expr.Equal, Value: expr.Int64Value(math.MinInt64),
			}},
		},
		{
			name:  "maximum uint64",
			input: "value = 18446744073709551615",
			want: filter.Filter{Expr: expr.ComparisonExpr{
				Field: "value", Op: expr.Equal, Value: expr.Uint64Value(math.MaxUint64),
			}},
		},
		{
			name:  "escaped unicode string",
			input: `name = "line\n\u4e16\u754c"`,
			want: filter.Filter{Expr: expr.ComparisonExpr{
				Field: "name", Op: expr.Equal, Value: expr.StringValue("line\n世界"),
			}},
		},
		{
			name:  "escaped quote",
			input: `"a\"b"`,
			want:  filter.Filter{Expr: expr.TextExpr{V: expr.StringValue(`a"b`)}},
		},
		{
			name:  "single quoted apostrophe",
			input: `'it\'s'`,
			want:  filter.Filter{Expr: expr.TextExpr{V: expr.StringValue("it's")}},
		},
		{
			name:  "double quote inside single quotes",
			input: `'a"b'`,
			want:  filter.Filter{Expr: expr.TextExpr{V: expr.StringValue(`a"b`)}},
		},
		{
			name:  "quoted keyword is text",
			input: `"and"`,
			want:  filter.Filter{Expr: expr.TextExpr{V: expr.StringValue("and")}},
		},
		{
			name:  "aip precedence",
			input: "a = 1 OR b = 2 AND c = 3",
			want: filter.Filter{Expr: expr.LogicalExpr{
				Left: expr.LogicalExpr{
					Left: expr.ComparisonExpr{
						Field: "a", Op: expr.Equal, Value: expr.Int64Value(1),
					},
					Op: expr.Or,
					Right: expr.ComparisonExpr{
						Field: "b", Op: expr.Equal, Value: expr.Int64Value(2),
					},
				},
				Op: expr.And,
				Right: expr.ComparisonExpr{
					Field: "c", Op: expr.Equal, Value: expr.Int64Value(3),
				},
			}},
		},
		{
			name:  "sequence",
			input: `"blue" "chair"`,
			want: filter.Filter{Expr: expr.LogicalExpr{
				Left:  expr.TextExpr{V: expr.StringValue("blue")},
				Op:    expr.Seq,
				Right: expr.TextExpr{V: expr.StringValue("chair")},
			}},
		},
		{
			name:  "word text",
			input: "chair",
			want:  filter.Filter{Expr: expr.TextExpr{V: expr.StringValue("chair")}},
		},
		{
			name:  "negation",
			input: "-deleted = true",
			want: filter.Filter{Expr: expr.NotExpr{
				Expr: expr.ComparisonExpr{
					Field: "deleted", Op: expr.Equal, Value: expr.BoolValue(true),
				},
			}},
		},
		{
			name:  "not and literal values",
			input: "NOT (active = false OR deleted = null)",
			want: filter.Filter{Expr: expr.NotExpr{Expr: expr.LogicalExpr{
				Left:  expr.ComparisonExpr{Field: "active", Op: expr.Equal, Value: expr.BoolValue(false)},
				Op:    expr.Or,
				Right: expr.ComparisonExpr{Field: "deleted", Op: expr.Equal, Value: expr.NullValue()},
			}}},
		},
		{
			name:  "field value",
			input: "state = active",
			want: filter.Filter{Expr: expr.ComparisonExpr{
				Field: "state", Op: expr.Equal, Value: expr.StringValue("active"),
			}},
		},
		{
			name:  "parenthesized",
			input: "(a = 1)",
			want:  filter.Filter{Expr: expr.ComparisonExpr{Field: "a", Op: expr.Equal, Value: expr.Int64Value(1)}},
		},
		{name: "trailing whitespace", input: "a ", want: filter.Filter{Expr: expr.TextExpr{V: expr.StringValue("a")}}},
		{
			name:  "terminal wildcard",
			input: `labels.* = "value"`,
			want: filter.Filter{Expr: expr.ComparisonExpr{
				Field: "labels.*", Op: expr.Equal, Value: expr.StringValue("value"),
			}},
		},
		{
			name:  "decimal without fraction",
			input: "1.",
			want:  filter.Filter{Expr: expr.TextExpr{V: expr.Float64Value(1)}},
		},
		{name: "positive exponent", input: "1e+2", want: filter.Filter{Expr: expr.TextExpr{V: expr.Float64Value(100)}}},
		{
			name:  "negative exponent",
			input: "1E-2",
			want:  filter.Filter{Expr: expr.TextExpr{V: expr.Float64Value(0.01)}},
		},
		{
			name:  "has",
			input: `tags:"blue"`,
			want: filter.Filter{Expr: expr.ComparisonExpr{
				Field: "tags", Op: expr.Has, Value: expr.StringValue("blue"),
			}},
		},
		{
			name:  "has bare identifier",
			input: "tags:blue",
			want: filter.Filter{Expr: expr.ComparisonExpr{
				Field: "tags", Op: expr.Has, Value: expr.StringValue("blue"),
			}},
		},
		{
			name:  "has dotted path",
			input: `metadata.tags:"blue"`,
			want: filter.Filter{Expr: expr.ComparisonExpr{
				Field: "metadata.tags", Op: expr.Has, Value: expr.StringValue("blue"),
			}},
		},
		{
			name:  "negated has",
			input: `NOT tags:"blue"`,
			want: filter.Filter{Expr: expr.NotExpr{Expr: expr.ComparisonExpr{
				Field: "tags", Op: expr.Has, Value: expr.StringValue("blue"),
			}}},
		},
		{
			name:  "googlesql escape",
			input: `'a\?b'`,
			want:  filter.Filter{Expr: expr.TextExpr{V: expr.StringValue("a?b")}},
		},
		{name: "double negation is not a term", input: `NOT NOT name = "x"`, wantErr: true},
		{name: "lowercase and", input: "a = 1 and b = 2", wantErr: true},
		{name: "lowercase or", input: "a = 1 or b = 2", wantErr: true},
		{name: "lowercase not", input: "not a = 1", wantErr: true},
		{name: "miscased keyword as argument", input: "provider = and", wantErr: true},
		{name: "function unsupported", input: "distance(location, point) < 10", wantErr: true},
		{name: "invalid encoding", input: string([]byte{0xff}), wantErr: true},
		{name: "invalid character", input: "@", wantErr: true},
		{name: "invalid operator", input: "active ! true", wantErr: true},
		{name: "invalid string escape", input: `name = "\q"`, wantErr: true},
		{name: "comparison without field", input: "1 = value", wantErr: true},
		{name: "malformed number", input: "value = 1e", wantErr: true},
		{name: "characters after numeric word", input: "1s+", wantErr: true},
		{name: "negative integer overflow", input: "value = -9223372036854775809", wantErr: true},
		{name: "unsigned integer overflow", input: "value = 18446744073709551616", wantErr: true},
		{name: "float overflow", input: "value = 1e999", wantErr: true},
		{name: "unterminated string", input: `name = "book`, wantErr: true},
		{name: "unbalanced composite", input: "(a = 1", wantErr: true},
		{name: "empty composite", input: "()", wantErr: true},
		{name: "not without whitespace", input: "NOT(a = 1)", wantErr: true},
		{name: "not without operand", input: "NOT ", wantErr: true},
		{name: "minus without operand", input: "-", wantErr: true},
		{name: "and without right operand", input: "a = 1 AND ", wantErr: true},
		{name: "and without trailing whitespace", input: "a AND", wantErr: true},
		{name: "and before invalid operand", input: "a = 1 AND )", wantErr: true},
		{name: "sequence before invalid operand", input: "a (", wantErr: true},
		{name: "or without right operand", input: "a = 1 OR ", wantErr: true},
		{name: "or without trailing whitespace", input: "a OR", wantErr: true},
		{name: "or before invalid operand", input: "a = 1 OR )", wantErr: true},
		{name: "unexpected closing parenthesis", input: "a)", wantErr: true},
		{name: "invalid traversal", input: "a. = 1", wantErr: true},
		{name: "wildcard must end path", input: "labels.*.value = 1", wantErr: true},
		{name: "invalid wildcard identifier", input: "labels** = 1", wantErr: true},
		{name: "nesting too deep", input: strings.Repeat("(", 65) + "a" + strings.Repeat(")", 65), wantErr: true},
		{name: "filter too large", input: strings.Repeat("a", (64<<10)+1), wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := filter.Parse(test.input)
			assert.Equal(t, test.want, got)
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
