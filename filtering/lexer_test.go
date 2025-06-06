package filtering

import (
	"errors"
	"io"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestLexer(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		filter        string
		expected      []Token
		errorContains string
	}{
		{
			filter: `New York Giants`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "New"},
				{Position: Position{Offset: 3, Column: 4, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeText, Value: "York"},
				{Position: Position{Offset: 8, Column: 9, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 9, Column: 10, Line: 1}, Type: TokenTypeText, Value: "Giants"},
			},
		},

		{
			filter: `New York Giants OR Yankees`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "New"},
				{Position: Position{Offset: 3, Column: 4, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeText, Value: "York"},
				{Position: Position{Offset: 8, Column: 9, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 9, Column: 10, Line: 1}, Type: TokenTypeText, Value: "Giants"},
				{Position: Position{Offset: 15, Column: 16, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 16, Column: 17, Line: 1}, Type: TokenTypeOr, Value: "OR"},
				{Position: Position{Offset: 18, Column: 19, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 19, Column: 20, Line: 1}, Type: TokenTypeText, Value: "Yankees"},
			},
		},

		{
			filter: `New York (Giants OR Yankees)`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "New"},
				{Position: Position{Offset: 3, Column: 4, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeText, Value: "York"},
				{Position: Position{Offset: 8, Column: 9, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 9, Column: 10, Line: 1}, Type: TokenTypeLeftParen, Value: "("},
				{Position: Position{Offset: 10, Column: 11, Line: 1}, Type: TokenTypeText, Value: "Giants"},
				{Position: Position{Offset: 16, Column: 17, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 17, Column: 18, Line: 1}, Type: TokenTypeOr, Value: "OR"},
				{Position: Position{Offset: 19, Column: 20, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 20, Column: 21, Line: 1}, Type: TokenTypeText, Value: "Yankees"},
				{Position: Position{Offset: 27, Column: 28, Line: 1}, Type: TokenTypeRightParen, Value: ")"},
			},
		},

		{
			filter: `a b AND c AND d`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "a"},
				{Position: Position{Offset: 1, Column: 2, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 2, Column: 3, Line: 1}, Type: TokenTypeText, Value: "b"},
				{Position: Position{Offset: 3, Column: 4, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeAnd, Value: "AND"},
				{Position: Position{Offset: 7, Column: 8, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 8, Column: 9, Line: 1}, Type: TokenTypeText, Value: "c"},
				{Position: Position{Offset: 9, Column: 10, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 10, Column: 11, Line: 1}, Type: TokenTypeAnd, Value: "AND"},
				{Position: Position{Offset: 13, Column: 14, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 14, Column: 15, Line: 1}, Type: TokenTypeText, Value: "d"},
			},
		},

		{
			filter: `(a b) AND c AND d`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeLeftParen, Value: "("},
				{Position: Position{Offset: 1, Column: 2, Line: 1}, Type: TokenTypeText, Value: "a"},
				{Position: Position{Offset: 2, Column: 3, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 3, Column: 4, Line: 1}, Type: TokenTypeText, Value: "b"},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeRightParen, Value: ")"},
				{Position: Position{Offset: 5, Column: 6, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 6, Column: 7, Line: 1}, Type: TokenTypeAnd, Value: "AND"},
				{Position: Position{Offset: 9, Column: 10, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 10, Column: 11, Line: 1}, Type: TokenTypeText, Value: "c"},
				{Position: Position{Offset: 11, Column: 12, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 12, Column: 13, Line: 1}, Type: TokenTypeAnd, Value: "AND"},
				{Position: Position{Offset: 15, Column: 16, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 16, Column: 17, Line: 1}, Type: TokenTypeText, Value: "d"},
			},
		},

		{
			filter: `a < 10 OR a >= 100`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "a"},
				{Position: Position{Offset: 1, Column: 2, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 2, Column: 3, Line: 1}, Type: TokenTypeLessThan, Value: "<"},
				{Position: Position{Offset: 3, Column: 4, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeNumber, Value: "10"},
				{Position: Position{Offset: 6, Column: 7, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 7, Column: 8, Line: 1}, Type: TokenTypeOr, Value: "OR"},
				{Position: Position{Offset: 9, Column: 10, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 10, Column: 11, Line: 1}, Type: TokenTypeText, Value: "a"},
				{Position: Position{Offset: 11, Column: 12, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 12, Column: 13, Line: 1}, Type: TokenTypeGreaterEquals, Value: ">="},
				{Position: Position{Offset: 14, Column: 15, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 15, Column: 16, Line: 1}, Type: TokenTypeNumber, Value: "100"},
			},
		},

		{
			filter: `NOT (a OR b)`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeNot, Value: "NOT"},
				{Position: Position{Offset: 3, Column: 4, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeLeftParen, Value: "("},
				{Position: Position{Offset: 5, Column: 6, Line: 1}, Type: TokenTypeText, Value: "a"},
				{Position: Position{Offset: 6, Column: 7, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 7, Column: 8, Line: 1}, Type: TokenTypeOr, Value: "OR"},
				{Position: Position{Offset: 9, Column: 10, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 10, Column: 11, Line: 1}, Type: TokenTypeText, Value: "b"},
				{Position: Position{Offset: 11, Column: 12, Line: 1}, Type: TokenTypeRightParen, Value: ")"},
			},
		},

		{
			filter: `-file:".java"`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeMinus, Value: "-"},
				{Position: Position{Offset: 1, Column: 2, Line: 1}, Type: TokenTypeText, Value: "file"},
				{Position: Position{Offset: 5, Column: 6, Line: 1}, Type: TokenTypeHas, Value: ":"},
				{Position: Position{Offset: 6, Column: 7, Line: 1}, Type: TokenTypeString, Value: `".java"`},
			},
		},

		{
			filter: `-30`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeMinus, Value: "-"},
				{Position: Position{Offset: 1, Column: 2, Line: 1}, Type: TokenTypeNumber, Value: "30"},
			},
		},

		{
			filter: `package=com.google`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "package"},
				{Position: Position{Offset: 7, Column: 8, Line: 1}, Type: TokenTypeEquals, Value: "="},
				{Position: Position{Offset: 8, Column: 9, Line: 1}, Type: TokenTypeText, Value: "com"},
				{Position: Position{Offset: 11, Column: 12, Line: 1}, Type: TokenTypeDot, Value: "."},
				{Position: Position{Offset: 12, Column: 13, Line: 1}, Type: TokenTypeText, Value: "google"},
			},
		},

		{
			filter: `msg != 'hello'`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "msg"},
				{Position: Position{Offset: 3, Column: 4, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeNotEquals, Value: "!="},
				{Position: Position{Offset: 6, Column: 7, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 7, Column: 8, Line: 1}, Type: TokenTypeString, Value: "'hello'"},
			},
		},

		{
			filter: `1 > 0`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeNumber, Value: "1"},
				{Position: Position{Offset: 1, Column: 2, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 2, Column: 3, Line: 1}, Type: TokenTypeGreaterThan, Value: ">"},
				{Position: Position{Offset: 3, Column: 4, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeNumber, Value: "0"},
			},
		},

		{
			filter: `2.5 >= 2.4`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeNumber, Value: "2"},
				{Position: Position{Offset: 1, Column: 2, Line: 1}, Type: TokenTypeDot, Value: "."},
				{Position: Position{Offset: 2, Column: 3, Line: 1}, Type: TokenTypeNumber, Value: "5"},
				{Position: Position{Offset: 3, Column: 4, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeGreaterEquals, Value: ">="},
				{Position: Position{Offset: 6, Column: 7, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 7, Column: 8, Line: 1}, Type: TokenTypeNumber, Value: "2"},
				{Position: Position{Offset: 8, Column: 9, Line: 1}, Type: TokenTypeDot, Value: "."},
				{Position: Position{Offset: 9, Column: 10, Line: 1}, Type: TokenTypeNumber, Value: "4"},
			},
		},

		{
			filter: `yesterday < request.time`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "yesterday"},
				{Position: Position{Offset: 9, Column: 10, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 10, Column: 11, Line: 1}, Type: TokenTypeLessThan, Value: "<"},
				{Position: Position{Offset: 11, Column: 12, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 12, Column: 13, Line: 1}, Type: TokenTypeText, Value: "request"},
				{Position: Position{Offset: 19, Column: 20, Line: 1}, Type: TokenTypeDot, Value: "."},
				{Position: Position{Offset: 20, Column: 21, Line: 1}, Type: TokenTypeText, Value: "time"},
			},
		},

		{
			filter: `experiment.rollout <= cohort(request.user)`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "experiment"},
				{Position: Position{Offset: 10, Column: 11, Line: 1}, Type: TokenTypeDot, Value: "."},
				{Position: Position{Offset: 11, Column: 12, Line: 1}, Type: TokenTypeText, Value: "rollout"},
				{Position: Position{Offset: 18, Column: 19, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 19, Column: 20, Line: 1}, Type: TokenTypeLessEquals, Value: "<="},
				{Position: Position{Offset: 21, Column: 22, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 22, Column: 23, Line: 1}, Type: TokenTypeText, Value: "cohort"},
				{Position: Position{Offset: 28, Column: 29, Line: 1}, Type: TokenTypeLeftParen, Value: "("},
				{Position: Position{Offset: 29, Column: 30, Line: 1}, Type: TokenTypeText, Value: "request"},
				{Position: Position{Offset: 36, Column: 37, Line: 1}, Type: TokenTypeDot, Value: "."},
				{Position: Position{Offset: 37, Column: 38, Line: 1}, Type: TokenTypeText, Value: "user"},
				{Position: Position{Offset: 41, Column: 42, Line: 1}, Type: TokenTypeRightParen, Value: ")"},
			},
		},

		{
			filter: `prod`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "prod"},
			},
		},

		{
			filter: `expr.type_map.1.type`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "expr"},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeDot, Value: "."},
				{Position: Position{Offset: 5, Column: 6, Line: 1}, Type: TokenTypeText, Value: "type_map"},
				{Position: Position{Offset: 13, Column: 14, Line: 1}, Type: TokenTypeDot, Value: "."},
				{Position: Position{Offset: 14, Column: 15, Line: 1}, Type: TokenTypeNumber, Value: "1"},
				{Position: Position{Offset: 15, Column: 16, Line: 1}, Type: TokenTypeDot, Value: "."},
				{Position: Position{Offset: 16, Column: 17, Line: 1}, Type: TokenTypeText, Value: "type"},
			},
		},

		{
			filter: `regex(m.key, '^.*prod.*$')`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "regex"},
				{Position: Position{Offset: 5, Column: 6, Line: 1}, Type: TokenTypeLeftParen, Value: "("},
				{Position: Position{Offset: 6, Column: 7, Line: 1}, Type: TokenTypeText, Value: "m"},
				{Position: Position{Offset: 7, Column: 8, Line: 1}, Type: TokenTypeDot, Value: "."},
				{Position: Position{Offset: 8, Column: 9, Line: 1}, Type: TokenTypeText, Value: "key"},
				{Position: Position{Offset: 11, Column: 12, Line: 1}, Type: TokenTypeComma, Value: ","},
				{Position: Position{Offset: 12, Column: 13, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 13, Column: 14, Line: 1}, Type: TokenTypeString, Value: "'^.*prod.*$'"},
				{Position: Position{Offset: 25, Column: 26, Line: 1}, Type: TokenTypeRightParen, Value: ")"},
			},
		},

		{
			filter: `math.mem('30mb')`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "math"},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeDot, Value: "."},
				{Position: Position{Offset: 5, Column: 6, Line: 1}, Type: TokenTypeText, Value: "mem"},
				{Position: Position{Offset: 8, Column: 9, Line: 1}, Type: TokenTypeLeftParen, Value: "("},
				{Position: Position{Offset: 9, Column: 10, Line: 1}, Type: TokenTypeString, Value: "'30mb'"},
				{Position: Position{Offset: 15, Column: 16, Line: 1}, Type: TokenTypeRightParen, Value: ")"},
			},
		},

		{
			filter: `(msg.endsWith('world') AND retries < 10)`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeLeftParen, Value: "("},
				{Position: Position{Offset: 1, Column: 2, Line: 1}, Type: TokenTypeText, Value: "msg"},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeDot, Value: "."},
				{Position: Position{Offset: 5, Column: 6, Line: 1}, Type: TokenTypeText, Value: "endsWith"},
				{Position: Position{Offset: 13, Column: 14, Line: 1}, Type: TokenTypeLeftParen, Value: "("},
				{Position: Position{Offset: 14, Column: 15, Line: 1}, Type: TokenTypeString, Value: "'world'"},
				{Position: Position{Offset: 21, Column: 22, Line: 1}, Type: TokenTypeRightParen, Value: ")"},
				{Position: Position{Offset: 22, Column: 23, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 23, Column: 24, Line: 1}, Type: TokenTypeAnd, Value: "AND"},
				{Position: Position{Offset: 26, Column: 27, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 27, Column: 28, Line: 1}, Type: TokenTypeText, Value: "retries"},
				{Position: Position{Offset: 34, Column: 35, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 35, Column: 36, Line: 1}, Type: TokenTypeLessThan, Value: "<"},
				{Position: Position{Offset: 36, Column: 37, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 37, Column: 38, Line: 1}, Type: TokenTypeNumber, Value: "10"},
				{Position: Position{Offset: 39, Column: 40, Line: 1}, Type: TokenTypeRightParen, Value: ")"},
			},
		},

		{
			filter: `foo = 0xdeadbeef`,
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "foo"},
				{Position: Position{Offset: 3, Column: 4, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 4, Column: 5, Line: 1}, Type: TokenTypeEquals, Value: "="},
				{Position: Position{Offset: 5, Column: 6, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 6, Column: 7, Line: 1}, Type: TokenTypeHexNumber, Value: "0xdeadbeef"},
			},
		},

		{
			filter:        `a = "foo`,
			errorContains: "unterminated string",
		},

		{
			filter:        "invalid = foo\xa0\x01bar",
			errorContains: "invalid UTF-8",
		},
		{
			filter: `object_id = "�g/ml" OR object_id = "µg/ml"`, // replacement character is valid UTF-8
			expected: []Token{
				{Position: Position{Offset: 0, Column: 1, Line: 1}, Type: TokenTypeText, Value: "object_id"},
				{Position: Position{Offset: 9, Column: 10, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 10, Column: 11, Line: 1}, Type: TokenTypeEquals, Value: "="},
				{Position: Position{Offset: 11, Column: 12, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 12, Column: 13, Line: 1}, Type: TokenTypeString, Value: `"�g/ml"`},
				{Position: Position{Offset: 21, Column: 20, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 22, Column: 21, Line: 1}, Type: TokenTypeOr, Value: "OR"},
				{Position: Position{Offset: 24, Column: 23, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 25, Column: 24, Line: 1}, Type: TokenTypeText, Value: "object_id"},
				{Position: Position{Offset: 34, Column: 33, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 35, Column: 34, Line: 1}, Type: TokenTypeEquals, Value: "="},
				{Position: Position{Offset: 36, Column: 35, Line: 1}, Type: TokenTypeWhitespace, Value: " "},
				{Position: Position{Offset: 37, Column: 36, Line: 1}, Type: TokenTypeString, Value: `"µg/ml"`},
			},
		},
	} {
		tt := tt
		t.Run(tt.filter, func(t *testing.T) {
			t.Parallel()
			var lexer Lexer
			lexer.Init(tt.filter)
			actual := make([]Token, 0, len(tt.expected))
			var tokenValues strings.Builder
			tokenValues.Grow(len(tt.filter))
			var token Token
			var err error
			for {
				token, err = lexer.Lex()
				if err != nil {
					break
				}
				_, _ = tokenValues.WriteString(token.Value)
				actual = append(actual, token)
			}
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.Assert(t, errors.Is(err, io.EOF))
				assert.DeepEqual(t, tt.expected, actual)
				assert.Equal(
					t,
					tt.filter,
					tokenValues.String(),
					"concatenating all token values should give the original value",
				)
			}
		})
	}
}

//nolint:gochecknoglobals
var tokenSink Token

func BenchmarkLexer_Lex(b *testing.B) {
	const filter = `(msg.endsWith('world') AND retries < 10)`
	b.ReportAllocs()
	var lexer Lexer
	for i := 0; i < b.N; i++ {
		lexer.Init(filter)
		token, _ := lexer.Lex()
		tokenSink = token
	}
}
