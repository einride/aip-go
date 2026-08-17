package flop

import (
	"errors"
	"io"
	"testing"

	"github.com/ysomad/flop/internal/assert"
)

// lexAll drains the lexer, returning every token before io.EOF.
func lexAll(t *testing.T, input string) ([]token, error) {
	t.Helper()
	l := newLexer(input)
	var tokens []token
	for {
		tok, err := l.lex()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tokens, nil
			}
			return tokens, err
		}
		tokens = append(tokens, tok)
	}
}

func types(tokens []token) []tokenType {
	kinds := make([]tokenType, 0, len(tokens))
	for _, tok := range tokens {
		kinds = append(kinds, tok.Type)
	}
	return kinds
}

func TestLexer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []tokenType
	}{
		{name: "comparison", input: "a = 1", want: []tokenType{tokenText, tokenWhitespace, tokenEquals, tokenWhitespace, tokenNumber}},
		{name: "two character operators", input: "a>=b", want: []tokenType{tokenText, tokenGreaterEquals, tokenText}},
		{name: "not equals", input: "a!=b", want: []tokenType{tokenText, tokenNotEquals, tokenText}},
		{name: "lone exclaim", input: "a!b", want: []tokenType{tokenText, tokenExclaim, tokenText}},
		{name: "has", input: `tags:"x"`, want: []tokenType{tokenText, tokenHas, tokenString}},
		{name: "keywords", input: "a AND b OR NOT c", want: []tokenType{
			tokenText, tokenWhitespace, tokenAnd, tokenWhitespace, tokenText, tokenWhitespace,
			tokenOr, tokenWhitespace, tokenNot, tokenWhitespace, tokenText,
		}},
		{name: "literals", input: "true false null", want: []tokenType{
			tokenTrue, tokenWhitespace, tokenFalse, tokenWhitespace, tokenNull,
		}},
		{name: "member", input: "a.b.c", want: []tokenType{tokenText, tokenDot, tokenText, tokenDot, tokenText}},
		{name: "wildcard", input: "labels.*", want: []tokenType{tokenText, tokenDot, tokenText}},
		{name: "signed number", input: "-1", want: []tokenType{tokenNumber}},
		{name: "minus before text", input: "-a", want: []tokenType{tokenMinus, tokenText}},
		{name: "float", input: "1.5", want: []tokenType{tokenNumber}},
		{name: "exponent", input: "1e+2", want: []tokenType{tokenNumber}},
		{name: "duration suffix folds into text", input: "2h", want: []tokenType{tokenText}},
		{name: "fractional duration folds into text", input: "1.5s", want: []tokenType{tokenText}},
		{name: "comma", input: "a,b", want: []tokenType{tokenText, tokenComma, tokenText}},
		{name: "single quoted string", input: `'x'`, want: []tokenType{tokenString}},
		{name: "escaped quote", input: `"a\"b"`, want: []tokenType{tokenString}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			tokens, err := lexAll(t, test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.want, types(tokens))
		})
	}
}

func TestLexer_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "unterminated double quote", input: `"abc`},
		{name: "unterminated single quote", input: `'abc`},
		{name: "invalid utf8", input: string([]byte{0xff})},
		{name: "invalid character", input: "@"},
		{name: "plus is not text", input: "+"},
		{name: "lowercase and", input: "a = 1 and b = 2"},
		{name: "mixed case or", input: "a = 1 Or b = 2"},
		{name: "lowercase not", input: "not a = 1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := lexAll(t, test.input)
			assert.Error(t, err)
			var filterErr *FilterError
			assert.ErrorAs(t, err, &filterErr)
		})
	}
}

func TestLexer_Positions(t *testing.T) {
	t.Parallel()

	tokens, err := lexAll(t, "a = 1\nb = 2")
	assert.NoError(t, err)

	// The newline is whitespace, so the second restriction starts on line 2.
	last := tokens[len(tokens)-1]
	assert.Equal(t, tokenNumber, last.Type)
	assert.Equal(t, int32(2), last.Position.Line)
	assert.Equal(t, int32(5), last.Position.Column)
	assert.Equal(t, int32(10), last.Position.Offset)
}

func TestLexer_PreservesReplacementCharacter(t *testing.T) {
	t.Parallel()

	// A literal U+FFFD inside a string is content, not an encoding error.
	tokens, err := lexAll(t, "\"�\"")
	assert.NoError(t, err)
	assert.Equal(t, []tokenType{tokenString}, types(tokens))
}
