package filtering

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// Lexer is a filter expression lexer.
type Lexer struct {
	filter string
	// current state
	err          error
	currToken    Token
	currTokenEnd Position
	currRune     rune
	// backtracking
	prevRune     rune
	prevToken    Token
	prevTokenEnd Position
}

func (l *Lexer) Init(filter string) {
	*l = Lexer{
		filter:       filter,
		currToken:    Token{Position: Position{Offset: 0, Line: 1, Column: 1}},
		currTokenEnd: Position{Offset: 0, Line: 1, Column: 1},
	}
}

// Token returns the current token in the filter expression.
func (l *Lexer) Token() Token {
	return l.currToken
}

// Err returns any error that was encountered during lexing of the filter expression.
func (l *Lexer) Err() error {
	return l.err
}

// Lex advances the lexer to the next token in the filter expression.
func (l *Lexer) Lex() bool {
	l.currToken = Token{Position: l.currTokenEnd}
	l.prevTokenEnd = l.currTokenEnd
	l.prevToken = l.currToken
	if !l.readRune() {
		return false
	}
	switch l.currRune {
	// Single-character operator?
	case '(', ')', '-', '.', '=', ':', ',':
		return l.emit(TokenType(l.currToken.Value))
	// Two-character operator?
	case '<', '>', '!':
		if !l.readRune() || l.currRune == '=' {
			return l.emit(TokenType(l.currToken.Value))
		}
		// Not a two-character operator. Back up and emit single-character operator.
		l.unreadRune()
		return l.emit(TokenType(l.currToken.Value))
	// Read string?
	case '\'', '"':
		delimiter := l.currRune
		for l.readRune() && l.currRune != delimiter {
			// Read until the closing delimiter.
		}
		if l.currRune != delimiter {
			return l.errorf("unterminated string")
		}
		return l.emit(TokenTypeString)
	}
	// Read whitespace?
	if unicode.IsSpace(l.currRune) {
		for l.readRune() {
			if !unicode.IsSpace(l.currRune) {
				l.unreadRune()
				break
			}
		}
		return l.emit(TokenTypeWhitespace)
	}
	// Read text.
	for l.readRune() {
		if !isText(l.currRune) {
			l.unreadRune()
			break
		}
	}
	switch l.currToken.Value {
	case "NOT", "AND", "OR":
		return l.emit(TokenType(l.currToken.Value))
	default:
		return l.emit(TokenTypeText)
	}
}

func (l *Lexer) emit(t TokenType) bool {
	l.currToken.Type = t
	return true
}

func (l *Lexer) errorf(format string, a ...interface{}) bool {
	l.err = fmt.Errorf(format, a...)
	return false
}

func (l *Lexer) eof() bool {
	return l.currTokenEnd.Offset >= len(l.filter)
}

func (l *Lexer) readRune() bool {
	if l.err != nil || l.eof() {
		return false
	}
	r, n := utf8.DecodeRuneInString(l.filter[l.currTokenEnd.Offset:])
	switch {
	case n == 0:
		return false
	case r == utf8.RuneError:
		return l.errorf("invalid UTF-8")
	}
	// update backtracking
	l.prevRune = l.currRune
	l.prevToken = l.currToken
	l.prevTokenEnd = l.currTokenEnd
	// update current state
	l.currRune = r
	l.currTokenEnd.advance(r, n)
	l.currToken.Value = l.filter[l.currToken.Position.Offset:l.currTokenEnd.Offset]
	return n > 0
}

func (l *Lexer) unreadRune() {
	l.currTokenEnd = l.prevTokenEnd
	l.currToken = l.prevToken
}

func isText(r rune) bool {
	switch r {
	case '(', ')', '-', '.', '=', ':', '<', '>', '!', ',':
		return false
	}
	return !unicode.IsSpace(r)
}
