package flop_test

import (
	"math"
	"strings"
	"testing"

	"github.com/ysomad/flop"
	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/internal/assert"
)

func TestParseFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantFilter flop.Filter
		wantErr    assert.ErrorFunc
	}{
		{name: "empty", input: " \n\t", wantFilter: flop.Filter{}, wantErr: assert.NoError},
		{
			name:  "comparison",
			input: "price >= 12.5",
			wantFilter: flop.Filter{Expr: expr.ComparisonExpr{
				Field: "price", Op: expr.GreaterOrEqual, Value: expr.Float64Value(12.5),
			}},
			wantErr: assert.NoError,
		},
		{
			name:  "all comparison operators",
			input: "a != 1 AND b < 2 AND c <= 3 AND d > 4",
			wantFilter: flop.Filter{Expr: expr.LogicalExpr{
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
			wantErr: assert.NoError,
		},
		{
			name:  "minimum int64",
			input: "value = -9223372036854775808",
			wantFilter: flop.Filter{Expr: expr.ComparisonExpr{
				Field: "value", Op: expr.Equal, Value: expr.Int64Value(math.MinInt64),
			}},
			wantErr: assert.NoError,
		},
		{
			name:  "maximum uint64",
			input: "value = 18446744073709551615",
			wantFilter: flop.Filter{Expr: expr.ComparisonExpr{
				Field: "value", Op: expr.Equal, Value: expr.Uint64Value(math.MaxUint64),
			}},
			wantErr: assert.NoError,
		},
		{
			name:  "escaped unicode string",
			input: `name = "line\n\u4e16\u754c"`,
			wantFilter: flop.Filter{Expr: expr.ComparisonExpr{
				Field: "name", Op: expr.Equal, Value: expr.StringValue("line\n世界"),
			}},
			wantErr: assert.NoError,
		},
		{
			name:       "escaped quote",
			input:      `"a\"b"`,
			wantFilter: flop.Filter{Expr: expr.TextExpr{V: expr.StringValue(`a"b`)}},
			wantErr:    assert.NoError,
		},
		{
			name:       "single quoted apostrophe",
			input:      `'it\'s'`,
			wantFilter: flop.Filter{Expr: expr.TextExpr{V: expr.StringValue("it's")}},
			wantErr:    assert.NoError,
		},
		{
			name:       "double quote inside single quotes",
			input:      `'a"b'`,
			wantFilter: flop.Filter{Expr: expr.TextExpr{V: expr.StringValue(`a"b`)}},
			wantErr:    assert.NoError,
		},
		{
			name:       "quoted keyword is text",
			input:      `"and"`,
			wantFilter: flop.Filter{Expr: expr.TextExpr{V: expr.StringValue("and")}},
			wantErr:    assert.NoError,
		},
		{
			name:  "aip precedence",
			input: "a = 1 OR b = 2 AND c = 3",
			wantFilter: flop.Filter{Expr: expr.LogicalExpr{
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
			wantErr: assert.NoError,
		},
		{
			name:  "sequence",
			input: `"blue" "chair"`,
			wantFilter: flop.Filter{Expr: expr.LogicalExpr{
				Left:  expr.TextExpr{V: expr.StringValue("blue")},
				Op:    expr.Seq,
				Right: expr.TextExpr{V: expr.StringValue("chair")},
			}},
			wantErr: assert.NoError,
		},
		{
			name:       "word text",
			input:      "chair",
			wantFilter: flop.Filter{Expr: expr.TextExpr{V: expr.StringValue("chair")}},
			wantErr:    assert.NoError,
		},
		{
			name:  "negation",
			input: "-deleted = true",
			wantFilter: flop.Filter{Expr: expr.NotExpr{
				Expr: expr.ComparisonExpr{
					Field: "deleted", Op: expr.Equal, Value: expr.BoolValue(true),
				},
			}},
			wantErr: assert.NoError,
		},
		{
			name:  "not and literal values",
			input: "NOT (active = false OR deleted = null)",
			wantFilter: flop.Filter{Expr: expr.NotExpr{Expr: expr.LogicalExpr{
				Left:  expr.ComparisonExpr{Field: "active", Op: expr.Equal, Value: expr.BoolValue(false)},
				Op:    expr.Or,
				Right: expr.ComparisonExpr{Field: "deleted", Op: expr.Equal, Value: expr.NullValue()},
			}}},
			wantErr: assert.NoError,
		},
		{
			name:  "field value",
			input: "state = active",
			wantFilter: flop.Filter{Expr: expr.ComparisonExpr{
				Field: "state", Op: expr.Equal, Value: expr.StringValue("active"),
			}},
			wantErr: assert.NoError,
		},
		{
			name:       "parenthesized",
			input:      "(a = 1)",
			wantFilter: flop.Filter{Expr: expr.ComparisonExpr{Field: "a", Op: expr.Equal, Value: expr.Int64Value(1)}},
			wantErr:    assert.NoError,
		},
		{
			name: "trailing whitespace", input: "a ",
			wantFilter: flop.Filter{Expr: expr.TextExpr{V: expr.StringValue("a")}},
			wantErr:    assert.NoError,
		},
		{
			name:  "terminal wildcard",
			input: `labels.* = "value"`,
			wantFilter: flop.Filter{Expr: expr.ComparisonExpr{
				Field: "labels.*", Op: expr.Equal, Value: expr.StringValue("value"),
			}},
			wantErr: assert.NoError,
		},
		{
			name:       "decimal without fraction",
			input:      "1.",
			wantFilter: flop.Filter{Expr: expr.TextExpr{V: expr.Float64Value(1)}},
			wantErr:    assert.NoError,
		},
		{
			name: "positive exponent", input: "1e+2",
			wantFilter: flop.Filter{Expr: expr.TextExpr{V: expr.Float64Value(100)}},
			wantErr:    assert.NoError,
		},
		{
			name:       "negative exponent",
			input:      "1E-2",
			wantFilter: flop.Filter{Expr: expr.TextExpr{V: expr.Float64Value(0.01)}},
			wantErr:    assert.NoError,
		},
		{
			name:  "has",
			input: `tags:"blue"`,
			wantFilter: flop.Filter{Expr: expr.ComparisonExpr{
				Field: "tags", Op: expr.Has, Value: expr.StringValue("blue"),
			}},
			wantErr: assert.NoError,
		},
		{
			name:  "has bare identifier",
			input: "tags:blue",
			wantFilter: flop.Filter{Expr: expr.ComparisonExpr{
				Field: "tags", Op: expr.Has, Value: expr.StringValue("blue"),
			}},
			wantErr: assert.NoError,
		},
		{
			name:  "has dotted path",
			input: `metadata.tags:"blue"`,
			wantFilter: flop.Filter{Expr: expr.ComparisonExpr{
				Field: "metadata.tags", Op: expr.Has, Value: expr.StringValue("blue"),
			}},
			wantErr: assert.NoError,
		},
		{
			name:  "negated has",
			input: `NOT tags:"blue"`,
			wantFilter: flop.Filter{Expr: expr.NotExpr{Expr: expr.ComparisonExpr{
				Field: "tags", Op: expr.Has, Value: expr.StringValue("blue"),
			}}},
			wantErr: assert.NoError,
		},
		{
			name:       "googlesql escape",
			input:      `'a\?b'`,
			wantFilter: flop.Filter{Expr: expr.TextExpr{V: expr.StringValue("a?b")}},
			wantErr:    assert.NoError,
		},
		{name: "double negation is not a term", input: `NOT NOT name = "x"`, wantErr: assert.Error},
		{name: "lowercase and", input: "a = 1 and b = 2", wantErr: assert.Error},
		{name: "lowercase or", input: "a = 1 or b = 2", wantErr: assert.Error},
		{name: "lowercase not", input: "not a = 1", wantErr: assert.Error},
		{name: "miscased keyword as argument", input: "provider = and", wantErr: assert.Error},
		{name: "function unsupported", input: "distance(location, point) < 10", wantErr: assert.Error},
		{name: "invalid encoding", input: string([]byte{0xff}), wantErr: assert.Error},
		{name: "invalid character", input: "@", wantErr: assert.Error},
		{name: "invalid operator", input: "active ! true", wantErr: assert.Error},
		{name: "invalid string escape", input: `name = "\q"`, wantErr: assert.Error},
		{name: "comparison without field", input: "1 = value", wantErr: assert.Error},
		{name: "malformed number", input: "value = 1e", wantErr: assert.Error},
		{name: "characters after numeric word", input: "1s+", wantErr: assert.Error},
		{name: "negative integer overflow", input: "value = -9223372036854775809", wantErr: assert.Error},
		{name: "unsigned integer overflow", input: "value = 18446744073709551616", wantErr: assert.Error},
		{name: "float overflow", input: "value = 1e999", wantErr: assert.Error},
		{name: "unterminated string", input: `name = "book`, wantErr: assert.Error},
		{name: "unbalanced composite", input: "(a = 1", wantErr: assert.Error},
		{name: "empty composite", input: "()", wantErr: assert.Error},
		{name: "not without whitespace", input: "NOT(a = 1)", wantErr: assert.Error},
		{name: "not without operand", input: "NOT ", wantErr: assert.Error},
		{name: "minus without operand", input: "-", wantErr: assert.Error},
		{name: "and without right operand", input: "a = 1 AND ", wantErr: assert.Error},
		{name: "and without trailing whitespace", input: "a AND", wantErr: assert.Error},
		{name: "and before invalid operand", input: "a = 1 AND )", wantErr: assert.Error},
		{name: "sequence before invalid operand", input: "a (", wantErr: assert.Error},
		{name: "or without right operand", input: "a = 1 OR ", wantErr: assert.Error},
		{name: "or without trailing whitespace", input: "a OR", wantErr: assert.Error},
		{name: "or before invalid operand", input: "a = 1 OR )", wantErr: assert.Error},
		{name: "unexpected closing parenthesis", input: "a)", wantErr: assert.Error},
		{name: "invalid traversal", input: "a. = 1", wantErr: assert.Error},
		{name: "wildcard must end path", input: "labels.*.value = 1", wantErr: assert.Error},
		{name: "invalid wildcard identifier", input: "labels** = 1", wantErr: assert.Error},
		{
			name:    "nesting too deep",
			input:   strings.Repeat("(", 65) + "a" + strings.Repeat(")", 65),
			wantErr: assert.Error,
		},
		{name: "filter too large", input: strings.Repeat("a", (64<<10)+1), wantErr: assert.Error},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := flop.ParseFilter(test.input)
			assert.Equal(t, test.wantFilter, got)
			test.wantErr(t, err)
		})
	}
}
