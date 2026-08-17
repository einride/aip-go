package filter

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/internal/ident"
)

const (
	maxBytes = 64 << 10
	maxDepth = 64
	maxNodes = 256
)

// Filter is the parsed form of an AIP-160 filter.
//
// The zero value represents no filtering.
type Filter struct {
	Expr expr.Expr
}

// String returns a canonical representation of f.
func (f Filter) String() string { return expr.Canonical(f.Expr) }

// Parse parses an AIP-160 filter.
//
// Functions are deliberately unsupported.
func Parse(input string) (Filter, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return Filter{}, nil
	}
	if len(input) > maxBytes {
		return Filter{}, newError(Position{Line: 1, Column: 1}, nil, "filter is too large")
	}
	p := newParser(trimmed)
	return p.parse()
}

// parser is a recursive-descent parser for the AIP-160 grammar.
//
// Backtracking copies the parser by value, so every field must remain safely
// copyable; the lexer holds only its input string and cursor positions.
type parser struct {
	lexer lexer
	depth int
}

// newParser expects input that Parse has already trimmed.
func newParser(filter string) parser {
	return parser{lexer: newLexer(filter)}
}

func (p *parser) parse() (Filter, error) {
	e, err := p.parseExpression()
	if err != nil {
		return Filter{}, err
	}
	end := p.lexer.position()
	tok, err := p.lexer.lex()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return Filter{}, err
		}
		return Filter{Expr: e}, nil
	}
	if tok.Type == tokenComma {
		return Filter{}, p.errorf(tok.Position, "functions are unsupported")
	}
	return Filter{}, p.errorf(end, "unexpected trailing token %s", tok.Type)
}

// operand is one side of a restriction: either a dotted field path or a
// literal value.
type operand struct {
	path   string
	value  expr.Value
	isPath bool
	pos    Position
}

// expression returns the operand as an unqualified text-search term.
func (o operand) expression() expr.Expr {
	if o.isPath {
		return expr.TextExpr{V: expr.StringValue(o.path)}
	}
	return expr.TextExpr{V: o.value}
}

// argument returns the operand as a comparison value. A bare identifier on the
// right-hand side is a string, never a field reference.
func (o operand) argument() expr.Value {
	if o.isPath {
		return expr.StringValue(o.path)
	}
	return o.value
}

// parseExpression parses an Expression.
//
//	expression
//	  : sequence {WS AND WS sequence}
//	  ;
func (p *parser) parseExpression() (expr.Expr, error) {
	sequences := make([]expr.Expr, 0, 1)
	for {
		_ = p.eatTokens(tokenWhitespace)
		sequence, err := p.parseSequence()
		if err != nil {
			return nil, err
		}
		sequences = append(sequences, sequence)
		if err := p.eatTokens(tokenWhitespace, tokenAnd, tokenWhitespace); err != nil {
			break
		}
	}
	e := sequences[0]
	for _, sequence := range sequences[1:] {
		e = expr.LogicalExpr{Left: e, Op: expr.And, Right: sequence}
	}
	return e, nil
}

// parseSequence parses a Sequence.
//
//	sequence
//	  : factor {WS factor}
//	  ;
func (p *parser) parseSequence() (expr.Expr, error) {
	factors := make([]expr.Expr, 0, 2)
	for {
		factor, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		factors = append(factors, factor)
		if p.sniffTokens(tokenWhitespace, tokenAnd) {
			break
		}
		if err := p.eatTokens(tokenWhitespace); err != nil {
			break
		}
	}
	e := factors[0]
	for _, factor := range factors[1:] {
		e = expr.LogicalExpr{Left: e, Op: expr.Seq, Right: factor}
	}
	return e, nil
}

// parseFactor parses a Factor.
//
//	factor
//	  : term {WS OR WS term}
//	  ;
func (p *parser) parseFactor() (expr.Expr, error) {
	terms := make([]expr.Expr, 0, 2)
	for {
		term, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		terms = append(terms, term)
		if err := p.eatTokens(tokenWhitespace, tokenOr, tokenWhitespace); err != nil {
			break
		}
	}
	e := terms[0]
	for _, term := range terms[1:] {
		e = expr.LogicalExpr{Left: e, Op: expr.Or, Right: term}
	}
	return e, nil
}

// parseTerm parses a Term.
//
//	term
//	  : [(NOT WS | MINUS)] simple
//	  ;
func (p *parser) parseTerm() (expr.Expr, error) {
	var negated bool
	if err := p.eatTokens(tokenNot, tokenWhitespace); err == nil {
		negated = true
	} else if err := p.eatTokens(tokenMinus); err == nil {
		negated = true
	}
	simple, err := p.parseSimple()
	if err != nil {
		return nil, err
	}
	if negated {
		return expr.NotExpr{Expr: simple}, nil
	}
	return simple, nil
}

// parseSimple parses a Simple.
//
//	simple
//	  : restriction
//	  | composite
//	  ;
func (p *parser) parseSimple() (expr.Expr, error) {
	if p.sniffTokens(tokenLeftParen) {
		return p.parseComposite()
	}
	return p.parseRestriction()
}

// parseComposite parses a Composite.
//
//	composite
//	  : LPAREN expression RPAREN
//	  ;
func (p *parser) parseComposite() (expr.Expr, error) {
	start := p.lexer.position()
	if p.depth >= maxDepth {
		return nil, p.errorf(start, "nesting is too deep")
	}
	p.depth++
	defer func() { p.depth-- }()

	if err := p.eatTokens(tokenLeftParen); err != nil {
		return nil, err
	}
	_ = p.eatTokens(tokenWhitespace)
	e, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	_ = p.eatTokens(tokenWhitespace)
	if err := p.eatTokens(tokenRightParen); err != nil {
		return nil, p.errorf(p.lexer.position(), "missing closing parenthesis")
	}
	return e, nil
}

// parseRestriction parses a Restriction.
//
//	restriction
//	  : comparable [comparator arg]
//	  ;
func (p *parser) parseRestriction() (expr.Expr, error) {
	comparable, err := p.parseComparable()
	if err != nil {
		return nil, err
	}
	if !p.sniff(tokenType.isComparator) && !p.sniff(tokenWhitespace.test, tokenType.isComparator) {
		return comparable.expression(), nil
	}
	_ = p.eatTokens(tokenWhitespace)
	comparator, err := p.parseToken(tokenType.isComparator)
	if err != nil {
		return nil, err
	}
	_ = p.eatTokens(tokenWhitespace)
	arg, err := p.parseComparable()
	if err != nil {
		return nil, err
	}
	if !comparable.isPath || !validPath(comparable.path) {
		return nil, p.errorf(comparable.pos, "comparison requires a field")
	}
	operator, ok := comparisonOp(comparator.Type)
	if !ok {
		return nil, p.errorf(comparator.Position, "invalid comparison operator")
	}
	return expr.ComparisonExpr{
		Field: comparable.path,
		Op:    operator,
		Value: arg.argument(),
	}, nil
}

// parseComparable parses a Comparable.
//
//	comparable
//	  : member
//	  | number
//	  | literal
//	  ;
func (p *parser) parseComparable() (operand, error) {
	pos := p.lexer.position()
	tok, err := p.parseToken(tokenType.isValue)
	if err != nil {
		return operand{}, err
	}
	switch tok.Type {
	case tokenString:
		text, err := tok.unquote()
		if err != nil {
			return operand{}, p.wrapf(err, tok.Position, "invalid string")
		}
		return operand{value: expr.StringValue(text), pos: pos}, nil
	case tokenNumber:
		value, err := parseNumber(tok.Value)
		if err != nil {
			return operand{}, p.wrapf(err, tok.Position, "invalid number")
		}
		return operand{value: value, pos: pos}, nil
	case tokenTrue:
		return operand{value: expr.BoolValue(true), pos: pos}, nil
	case tokenFalse:
		return operand{value: expr.BoolValue(false), pos: pos}, nil
	case tokenNull:
		return operand{value: expr.NullValue(), pos: pos}, nil
	default:
		return p.parseMember(tok, pos)
	}
}

// parseMember parses a Member.
//
//	member
//	  : value {DOT field}
//	  ;
func (p *parser) parseMember(first token, pos Position) (operand, error) {
	parts := []string{first.Value}
	for {
		if err := p.eatTokens(tokenDot); err != nil {
			break
		}
		field, err := p.parseToken(tokenType.isField)
		if err != nil {
			return operand{}, p.errorf(field.Position, "invalid field traversal")
		}
		parts = append(parts, field.Value)
	}
	return operand{path: strings.Join(parts, "."), isPath: true, pos: pos}, nil
}

func (p *parser) parseToken(fns ...func(tokenType) bool) (token, error) {
	pos := p.lexer.position()
	tok, err := p.lexer.lex()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return token{}, p.errorf(pos, "unexpected end of filter")
		}
		return token{}, err
	}
	for _, fn := range fns {
		if fn(tok.Type) {
			return tok, nil
		}
	}
	if tok.Type == tokenComma {
		return token{}, p.errorf(tok.Position, "functions are unsupported")
	}
	return token{}, p.errorf(tok.Position, "unexpected token %s", tok.Type)
}

func (p *parser) sniff(fns ...func(tokenType) bool) bool {
	start := *p
	defer func() { *p = start }()
	for _, fn := range fns {
		tok, err := p.lexer.lex()
		if err != nil || !fn(tok.Type) {
			return false
		}
	}
	return true
}

func (p *parser) sniffTokens(want ...tokenType) bool {
	start := *p
	defer func() { *p = start }()
	for _, wantType := range want {
		if tok, err := p.lexer.lex(); err != nil || tok.Type != wantType {
			return false
		}
	}
	return true
}

func (p *parser) eatTokens(want ...tokenType) error {
	start := *p
	for _, wantType := range want {
		tok, err := p.lexer.lex()
		if err != nil || tok.Type != wantType {
			*p = start
			return p.errorf(start.lexer.position(), "expected %s", wantType)
		}
	}
	return nil
}

func (p *parser) errorf(pos Position, format string, args ...any) error {
	return newError(pos, nil, format, args...)
}

func (p *parser) wrapf(err error, pos Position, format string, args ...any) error {
	return newError(pos, err, format, args...)
}

func comparisonOp(t tokenType) (expr.ComparisonOp, bool) {
	switch t {
	case tokenEquals:
		return expr.Equal, true
	case tokenNotEquals:
		return expr.NotEqual, true
	case tokenLess:
		return expr.Less, true
	case tokenLessEquals:
		return expr.LessOrEqual, true
	case tokenGreater:
		return expr.Greater, true
	case tokenGreaterEquals:
		return expr.GreaterOrEqual, true
	case tokenHas:
		return expr.Has, true
	default:
		return 0, false
	}
}

func parseNumber(literal string) (expr.Value, error) {
	if strings.ContainsAny(literal, ".eE") {
		value, err := strconv.ParseFloat(literal, 64)
		if err != nil {
			return expr.Value{}, err
		}
		return expr.Float64Value(value), nil
	}
	if strings.HasPrefix(literal, "-") {
		value, err := strconv.ParseInt(literal, 10, 64)
		if err != nil {
			return expr.Value{}, err
		}
		return expr.Int64Value(value), nil
	}
	if value, err := strconv.ParseInt(literal, 10, 64); err == nil {
		return expr.Int64Value(value), nil
	}
	value, err := strconv.ParseUint(literal, 10, 64)
	if err != nil {
		return expr.Value{}, err
	}
	return expr.Uint64Value(value), nil
}

// validPath reports whether path is a filter field path: dot-separated
// identifiers, optionally ending in the AIP-160 wildcard.
func validPath(path string) bool {
	parts := strings.Split(path, ".")
	for i, part := range parts {
		if part == "*" {
			if i != len(parts)-1 {
				return false
			}
			continue
		}
		if !ident.Valid(part) {
			return false
		}
	}
	return true
}
