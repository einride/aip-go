package filtering

import "fmt"

// ParseExpression parses a filter expression.
func ParseExpression(filter string) (Expression, error) {
	var parser Parser
	parser.Init(filter)
	return parser.ParseExpression()
}

type Parser struct {
	filter string
	lexer  Lexer
}

func (p *Parser) Init(filter string) {
	p.filter = filter
	p.lexer.Init(filter)
}

// ParseExpression parses an Expression.
//
// EBNF
//
//  expression
//    : sequence {WS AND WS sequence}
//    ;
func (p *Parser) ParseExpression() (Expression, error) {
	var expression Expression
	for {
		sequence, err := p.ParseSequence()
		if err != nil {
			return expression, err
		}
		expression.Sequences = append(expression.Sequences, sequence)
		if err := p.eat(TokenTypeWhitespace, TokenTypeAnd, TokenTypeWhitespace); err != nil {
			break
		}
	}
	return expression, nil
}

// ParseSequence parses a Sequence.
//
// EBNF
//
//  sequence
//    : factor {WS factor}
//    ;
func (p *Parser) ParseSequence() (Sequence, error) {
	var sequence Sequence
	for {
		factor, err := p.ParseFactor()
		if err != nil {
			return sequence, err
		}
		sequence.Factors = append(sequence.Factors, factor)
		if err := p.eat(TokenTypeWhitespace); err != nil {
			break
		}
	}
	return sequence, nil
}

// ParseFactor parses a Factor.
//
// EBNF
//
//  factor
//    : term {WS OR WS term}
//    ;
func (p *Parser) ParseFactor() (Factor, error) {
	var factor Factor
	for {
		term, err := p.ParseTerm()
		if err != nil {
			return factor, err
		}
		factor.Terms = append(factor.Terms, term)
		if err := p.eat(TokenTypeWhitespace, TokenTypeOr, TokenTypeWhitespace); err != nil {
			break
		}
	}
	return factor, nil
}

// ParseTerm parses a Term.
//
// EBNF
//
//  term
//    : [(NOT WS | MINUS)] simple
//    ;
func (p *Parser) ParseTerm() (Term, error) {
	var term Term
	if err := p.eat(TokenTypeNot, TokenTypeWhitespace); err == nil {
		term.Not = true
	} else if err := p.eat(TokenTypeMinus); err == nil {
		term.Minus = true
	}
	simple, err := p.ParseSimple()
	if err != nil {
		return term, err
	}
	term.Simple = simple
	return term, nil
}

// ParseSimple parses a Simple.
//
// EBNF
//
//  simple
//    : restriction
//    | composite
//    ;
func (p *Parser) ParseSimple() (Simple, error) {
	if composite, ok := p.TryParseComposite(); ok {
		return composite, nil
	}
	return p.ParseRestriction()
}

// ParseRestriction parses a Restriction.
//
// EBNF
//
//  restriction
//    : comparable [comparator arg]
//    ;
func (p *Parser) ParseRestriction() (Restriction, error) {
	var restriction Restriction
	comparable, err := p.ParseComparable()
	if err != nil {
		return restriction, err
	}
	restriction.Comparable = comparable
	if comparator, ok := p.TryParseComparator(); ok {
		restriction.Comparator = comparator
		arg, err := p.ParseArg()
		if err != nil {
			return restriction, err
		}
		restriction.Arg = arg
	}
	return restriction, nil
}

// ParseComparable parses a Comparable.
//
// EBNF
//
//  comparable
//    : member
//    | function
//    ;
func (p *Parser) ParseComparable() (Comparable, error) {
	if function, ok := p.TryParseFunction(); ok {
		return function, nil
	}
	return p.ParseMember()
}

// ParseMember parses a Member.
//
// EBNF
//
//  member
//    : value {DOT field}
//    ;
func (p *Parser) ParseMember() (Member, error) {
	var member Member
	value, err := p.ParseValue()
	if err != nil {
		return member, err
	}
	member.Value = value
	for {
		if err := p.eat(TokenTypeDot); err != nil {
			break
		}
		field, err := p.ParseField()
		if err != nil {
			return member, err
		}
		member.Fields = append(member.Fields, field)
	}
	return member, nil
}

// ParseFunction parses a Function.
//
// EBNF
//
//  function
//    : name {DOT name} LPAREN [argList] RPAREN
//    ;
func (p *Parser) ParseFunction() (Function, error) {
	var function Function
	for {
		name, err := p.ParseName()
		if err != nil {
			return function, err
		}
		function.Names = append(function.Names, name)
		if err := p.eat(TokenTypeDot); err != nil {
			break
		}
	}
	if err := p.eat(TokenTypeLeftParen); err != nil {
		return function, err
	}
	for !p.sniff(TokenTypeRightParen) {
		arg, err := p.ParseArg()
		if err != nil {
			return function, err
		}
		function.Args = append(function.Args, arg)
		if err := p.eat(TokenTypeComma); err != nil {
			break
		}
	}
	if err := p.eat(TokenTypeRightParen); err != nil {
		return function, err
	}
	return function, nil
}

func (p *Parser) TryParseFunction() (Function, bool) {
	start := *p
	function, err := p.ParseFunction()
	if err != nil {
		*p = start
		return Function{}, false
	}
	return function, true
}

// ParseComparator parses a Comparator.
//
// EBNF
//
//  comparator
//    : LESS_EQUALS      # <=
//    | LESS_THAN        # <
//    | GREATER_EQUALS   # >=
//    | GREATER_THAN     # >
//    | NOT_EQUALS       # !=
//    | EQUALS           # =
//    | HAS              # :
//    ;
func (p *Parser) ParseComparator() (Comparator, error) {
	if !p.lexer.Lex() {
		return "", p.errorf("expected comparator, got EOF")
	}
	switch p.lexer.Token().Type {
	case TokenTypeLessEquals,
		TokenTypeLess,
		TokenTypeGreaterEquals,
		TokenTypeGreater,
		TokenTypeNotEquals,
		TokenTypeEquals,
		TokenTypeHas:
		return Comparator(p.lexer.Token().Value), nil
	default:
		return "", p.errorf("expected comparator, got %s", p.lexer.Token().Type)
	}
}

func (p *Parser) TryParseComparator() (Comparator, bool) {
	start := *p
	comparator, err := p.ParseComparator()
	if err != nil {
		*p = start
		return comparator, false
	}
	return comparator, true
}

// ParseComposite parses a Composite.
//
// EBNF
//
//  composite
//    : LPAREN expression RPAREN
//    ;
func (p *Parser) ParseComposite() (Composite, error) {
	var composite Composite
	if err := p.eat(TokenTypeLeftParen); err != nil {
		return composite, err
	}
	expression, err := p.ParseExpression()
	if err != nil {
		return composite, err
	}
	composite.Expression = expression
	if err := p.eat(TokenTypeRightParen); err != nil {
		return composite, err
	}
	return composite, nil
}

// TryParseComposite tries to parse Composite..
func (p *Parser) TryParseComposite() (Composite, bool) {
	start := *p
	composite, err := p.ParseComposite()
	if err != nil {
		*p = start
		return composite, false
	}
	return composite, true
}

// ParseValue parses a Value.
//
// EBNF
//
//  value
//    : TEXT
//    | STRING
//    ;
func (p *Parser) ParseValue() (Value, error) {
	if !p.lexer.Lex() {
		return nil, fmt.Errorf("expected value, got EOF")
	}
	switch p.lexer.Token().Type {
	case TokenTypeText:
		return Text(p.lexer.Token().Value), nil
	case TokenTypeString:
		return String(p.lexer.Token().Value), nil
	default:
		return nil, fmt.Errorf("expected value, got %s", p.lexer.Token().Type)
	}
}

// ParseField parses a Field.
//
// EBNF
//
//  field
//    : value
//    | keyword
//    ;
func (p *Parser) ParseField() (Field, error) {
	if keyword, ok := p.TryParseKeyword(); ok {
		return keyword, nil
	}
	return p.ParseValue()
}

// ParseName parses a Name.
//
// EBNF
//
//  name
//    : TEXT
//    | keyword
//    ;
func (p *Parser) ParseName() (Name, error) {
	if keyword, ok := p.TryParseKeyword(); ok {
		return keyword, nil
	}
	if !p.lexer.Lex() {
		return nil, p.errorf("expected name, got EOF")
	}
	if p.lexer.Token().Type != TokenTypeText {
		return nil, p.errorf("expected name, got %s", p.lexer.Token().Type)
	}
	return Text(p.lexer.Token().Value), nil
}

// ParseArg parses an Arg.
//
// EBNF
//
//  arg
//    : comparable
//    | composite
//    ;
func (p *Parser) ParseArg() (Arg, error) {
	if composite, ok := p.TryParseComposite(); ok {
		return composite, nil
	}
	return p.ParseComparable()
}

// ParseKeyword parses a Keyword.
//
// EBNF
//
//  keyword
//    : NOT
//    | AND
//    | OR
//    ;
func (p *Parser) ParseKeyword() (Keyword, error) {
	if !p.lexer.Lex() {
		return "", fmt.Errorf("expected keyword")
	}
	token := p.lexer.Token()
	switch token.Type {
	case TokenTypeNot, TokenTypeAnd, TokenTypeOr:
		return Keyword(token.Value), nil
	default:
		return "", p.errorf("expected keyword, got %s", token.Type)
	}
}

func (p *Parser) TryParseKeyword() (Keyword, bool) {
	start := *p
	keyword, err := p.ParseKeyword()
	if err != nil {
		*p = start
		return keyword, false
	}
	return keyword, true
}

func (p *Parser) sniff(wantTokenTypes ...TokenType) bool {
	start := *p
	defer func() {
		*p = start
	}()
	for _, wantTokenType := range wantTokenTypes {
		if !p.lexer.Lex() || p.lexer.Token().Type != wantTokenType {
			return false
		}
	}
	return true
}

func (p *Parser) eat(wantTokenTypes ...TokenType) error {
	start := *p
	for _, wantTokenType := range wantTokenTypes {
		if !p.lexer.Lex() || p.lexer.Token().Type != wantTokenType {
			*p = start
			return p.errorf("expected %s", wantTokenType)
		}
	}
	return nil
}

func (p *Parser) errorf(format string, args ...interface{}) error {
	return &ParseError{
		Filter:   p.filter,
		Position: p.lexer.Token().Position,
		Message:  fmt.Sprintf(format, args...),
	}
}

type ParseError struct {
	Filter   string
	Position Position
	Message  string
}

func (p *ParseError) Error() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%d:%d: %s", p.Position.Line, p.Position.Column, p.Message)
}

var _ error = &ParseError{}
