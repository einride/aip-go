package flop

import (
	"errors"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ysomad/flop/internal/ident"
)

// FilterPosition is a position in a filter expression.
type FilterPosition struct {
	Offset int32
	Line   int32
	Column int32
}

type tokenType string

const (
	tokenWhitespace tokenType = "WS"
	tokenText       tokenType = "TEXT"
	tokenString     tokenType = "STRING"
	tokenNumber     tokenType = "NUM"

	tokenNot tokenType = "NOT"
	tokenAnd tokenType = "AND"
	tokenOr  tokenType = "OR"

	tokenTrue  tokenType = "true"
	tokenFalse tokenType = "false"
	tokenNull  tokenType = "null"

	tokenLeftParen     tokenType = "("
	tokenRightParen    tokenType = ")"
	tokenMinus         tokenType = "-"
	tokenDot           tokenType = "."
	tokenComma         tokenType = ","
	tokenEquals        tokenType = "="
	tokenHas           tokenType = ":"
	tokenLess          tokenType = "<"
	tokenGreater       tokenType = ">"
	tokenExclaim       tokenType = "!"
	tokenLessEquals    tokenType = "<="
	tokenGreaterEquals tokenType = ">="
	tokenNotEquals     tokenType = "!="
)

// isKeyword reports whether t is a logical keyword.
func (t tokenType) isKeyword() bool {
	switch t {
	case tokenNot, tokenAnd, tokenOr:
		return true
	default:
		return false
	}
}

// miscasedKeyword reports whether text is a logical keyword in the wrong case.
// AIP-160 spells AND, OR, and NOT uppercase; treating a lowercase spelling as a
// text-search term silently changes what the filter means, so reject it and let
// a quoted string carry the literal word.
func miscasedKeyword(text string) bool {
	switch {
	case strings.EqualFold(text, string(tokenAnd)),
		strings.EqualFold(text, string(tokenOr)),
		strings.EqualFold(text, string(tokenNot)):
		return true
	default:
		return false
	}
}

// isLiteral reports whether t is a bare literal keyword.
func (t tokenType) isLiteral() bool {
	switch t {
	case tokenTrue, tokenFalse, tokenNull:
		return true
	default:
		return false
	}
}

// isValue reports whether t may begin a comparable.
func (t tokenType) isValue() bool {
	switch t {
	case tokenText, tokenString, tokenNumber:
		return true
	default:
		return t.isLiteral()
	}
}

// isField reports whether t may follow a dot in a member.
func (t tokenType) isField() bool { return t == tokenText }

func (t tokenType) isComparator() bool {
	switch t {
	case tokenEquals, tokenNotEquals, tokenLess, tokenLessEquals,
		tokenGreater, tokenGreaterEquals, tokenHas:
		return true
	default:
		return false
	}
}

func (t tokenType) test(other tokenType) bool { return t == other }

type token struct {
	Position FilterPosition
	Type     tokenType
	Value    string
}

// unquote returns the token's textual value, unescaping string literals.
func (t token) unquote() (string, error) {
	if t.Type == tokenString {
		return unescape(t.Value)
	}
	return t.Value, nil
}

// lexer is a filter expression lexer.
//
// It is copyable: the parser saves and restores one by value to backtrack.
type lexer struct {
	filter     string
	tokenStart FilterPosition
	tokenEnd   FilterPosition
}

func newLexer(filter string) lexer {
	start := FilterPosition{Offset: 0, Line: 1, Column: 1}
	return lexer{filter: filter, tokenStart: start, tokenEnd: start}
}

// lex returns the next token, or io.EOF when the input is exhausted.
func (l *lexer) lex() (token, error) {
	r, err := l.nextRune()
	if err != nil {
		return token{}, err
	}
	switch r {
	case '(', ')', '.', '=', ':', ',':
		return l.emit(tokenType(l.tokenValue()))
	case '<', '>', '!':
		if l.sniffRune('=') {
			_, _ = l.nextRune()
		}
		return l.emit(tokenType(l.tokenValue()))
	case '\'', '"':
		return l.lexString(r)
	case '-':
		if l.sniff(isDigit) {
			return l.lexNumber()
		}
		return l.emit(tokenMinus)
	}
	if isDigit(r) {
		return l.lexNumber()
	}
	if unicode.IsSpace(r) {
		for l.sniff(unicode.IsSpace) {
			_, _ = l.nextRune()
		}
		return l.emit(tokenWhitespace)
	}
	if !ident.IsStart(r) && r != '*' {
		return token{}, l.errorf("invalid character")
	}
	for l.sniff(isText) {
		_, _ = l.nextRune()
	}
	value := l.tokenValue()
	if t := tokenType(value); t.isKeyword() || t.isLiteral() {
		return l.emit(t)
	}
	if miscasedKeyword(value) {
		return token{}, l.errorf("logical operator %q must be uppercase", value)
	}
	return l.emit(tokenText)
}

func (l *lexer) lexString(quote rune) (token, error) {
	escaped := false
	for {
		r, err := l.nextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return token{}, l.errorf("unterminated string")
			}
			return token{}, err
		}
		if r == '\\' && !escaped {
			escaped = true
			continue
		}
		if r == quote && !escaped {
			return l.emit(tokenString)
		}
		escaped = false
	}
}

// lexNumber consumes a whole numeric literal: an optional sign, digits, an
// optional fraction, and an optional exponent.
func (l *lexer) lexNumber() (token, error) {
	for l.sniff(isDigit) {
		_, _ = l.nextRune()
	}
	if l.sniffRune('.') {
		_, _ = l.nextRune()
		for l.sniff(isDigit) {
			_, _ = l.nextRune()
		}
	}
	if l.sniffRune('e') || l.sniffRune('E') {
		_, _ = l.nextRune()
		if l.sniffRune('+') || l.sniffRune('-') {
			_, _ = l.nextRune()
		}
		for l.sniff(isDigit) {
			_, _ = l.nextRune()
		}
	}
	if !l.sniff(ident.IsStart) {
		return l.emit(tokenNumber)
	}
	for l.sniff(isText) {
		_, _ = l.nextRune()
	}
	return l.emit(tokenText)
}

func (l *lexer) position() FilterPosition { return l.tokenStart }

func (l *lexer) emit(t tokenType) (token, error) {
	tok := token{
		Position: l.tokenStart,
		Type:     t,
		Value:    l.tokenValue(),
	}
	l.tokenStart = l.tokenEnd
	return tok, nil
}

func (l *lexer) tokenValue() string {
	return l.filter[l.tokenStart.Offset:l.tokenEnd.Offset]
}

func (l *lexer) remaining() string {
	return l.filter[l.tokenEnd.Offset:]
}

func (l *lexer) nextRune() (rune, error) {
	r, n := utf8.DecodeRuneInString(l.remaining())
	switch {
	case n == 0:
		return r, io.EOF
	// A literal replacement character is preserved; genuinely invalid encoding
	// is rejected.
	case r == utf8.RuneError && (n != 3 || l.remaining()[:3] != "\xef\xbf\xbd"):
		return r, l.errorf("invalid UTF-8")
	}
	if r == '\n' {
		l.tokenEnd.Line++
		l.tokenEnd.Column = 1
	} else {
		l.tokenEnd.Column++
	}
	l.tokenEnd.Offset += int32(n)
	return r, nil
}

func (l *lexer) sniff(want func(rune) bool) bool {
	r, n := utf8.DecodeRuneInString(l.remaining())
	return n > 0 && want(r)
}

func (l *lexer) sniffRune(want rune) bool {
	r, _ := utf8.DecodeRuneInString(l.remaining())
	return r == want
}

func (l *lexer) errorf(format string, args ...any) error {
	return newError(l.tokenStart, nil, format, args...)
}

func isText(r rune) bool { return ident.IsContinue(r) || r == '*' }

func isDigit(r rune) bool { return r >= '0' && r <= '9' }
