package filtering

import "strings"

// Expression may either be a conjunction (AND) of sequences or a simple
// sequence.
//
// Note, the AND is case-sensitive.
//
// Example: `a b AND c AND d`
//
// The expression `(a b) AND c AND d` is equivalent to the example.
//
// EBNF
//
//  expression
//    : sequence {WS AND WS sequence}
//    ;
type Expression struct {
	// Sequences in the expression.
	Sequences []Sequence
}

func (e Expression) Filter() string {
	var sb strings.Builder
	e.WriteFilter(&sb)
	return sb.String()
}

func (e Expression) WriteFilter(sb *strings.Builder) {
	for i, sequence := range e.Sequences {
		sequence.WriteFilter(sb)
		if i < len(e.Sequences)-1 {
			sb.WriteString(" AND ")
		}
	}
}

// Sequence is composed of one or more whitespace (WS) separated factors.
//
// A sequence expresses a logical relationship between 'factors' where
// the ranking of a filter result may be scored according to the number
// factors that match and other such criteria as the proximity of factors
// to each other within a document.
//
// When filters are used with exact match semantics rather than fuzzy
// match semantics, a sequence is equivalent to AND.
//
// Example: `New York Giants OR Yankees`
//
// The expression `New York (Giants OR Yankees)` is equivalent to the
// example.
//
// EBNF
//
//  sequence
//    : factor {WS factor}
//    ;
type Sequence struct {
	// Factors in the sequence.
	Factors []Factor
}

func (s Sequence) Filter() string {
	var sb strings.Builder
	s.WriteFilter(&sb)
	return sb.String()
}

func (s Sequence) WriteFilter(sb *strings.Builder) {
	for i, factor := range s.Factors {
		factor.WriteFilter(sb)
		if i < len(s.Factors)-1 {
			_ = sb.WriteByte(' ')
		}
	}
}

// Factor may either be a disjunction (OR) of terms or a simple term.
//
// Note, the OR is case-sensitive.
//
// Example: `a < 10 OR a >= 100`
//
// EBNF
//
//  factor
//    : term {WS OR WS term}
//    ;
type Factor struct {
	// Terms of the factor.
	Terms []Term
}

func (f Factor) Filter() string {
	var sb strings.Builder
	f.WriteFilter(&sb)
	return sb.String()
}

func (f Factor) WriteFilter(sb *strings.Builder) {
	for i, term := range f.Terms {
		term.WriteFilter(sb)
		if i < len(f.Terms)-1 {
			_, _ = sb.WriteString(" OR ")
		}
	}
}

// Terms may either be unary or simple expressions.
//
// Unary expressions negate the simple expression, either mathematically `-`
// or logically `NOT`. The negation styles may be used interchangeably.
//
// Note, the `NOT` is case-sensitive and must be followed by at least one
// whitespace (WS).
//
// Examples:
// * logical not     : `NOT (a OR b)`
// * alternative not : `-file:".java"`
// * negation        : `-30`
//
// EBNF
//
//  term
//    : [(NOT WS | MINUS)] simple
//    ;
type Term struct {
	// Not indicates if the term is negated with NOT.
	Not bool
	// Minus indicates if the term is negated with MINUS.
	Minus bool
	// Simple expression.
	Simple Simple
}

func (t Term) Filter() string {
	var sb strings.Builder
	t.WriteFilter(&sb)
	return sb.String()
}

func (t Term) WriteFilter(sb *strings.Builder) {
	switch {
	case t.Not:
		_, _ = sb.WriteString(string(TokenTypeNot))
		_ = sb.WriteByte(' ')
	case t.Minus:
		_, _ = sb.WriteString(string(TokenTypeMinus))
	}
	t.Simple.WriteFilter(sb)
}

// Simple expressions may either be a restriction or a nested (composite)
// expression.
//
// EBNF
//
//  simple
//    : restriction
//    | composite
//    ;
type Simple interface {
	isSimple()
	WriteFilter(*strings.Builder)
}

var (
	_ Simple = Restriction{}
	_ Simple = Composite{}
)

// Restrictions express a relationship between a comparable value and a
// single argument. When the restriction only specifies a comparable
// without an operator, this is a global restriction.
//
// Note, restrictions are not whitespace sensitive.
//
// Examples
//
//  * equality         : `package=com.google`
//  * inequality       : `msg != 'hello'`
//  * greater than     : `1 > 0`
//  * greater or equal : `2.5 >= 2.4`
//  * less than        : `yesterday < request.time`
//  * less or equal    : `experiment.rollout <= cohort(request.user)`
//  * has              : `map:key`
//  * global           : `prod`
//
// In addition to the global, equality, and ordering operators, filters
// also support the has (`:`) operator. The has operator is unique in
// that it can test for presence or value based on the proto3 type of
// the `comparable` value. The has operator is useful for validating the
// structure and contents of complex values.
//
// EBNF
//
//  restriction
//    : comparable [comparator arg]
//    ;
type Restriction struct {
	// Comparable value.
	Comparable Comparable
	// Comparator of the restriction.
	Comparator Comparator
	// Arg of the restriction.
	Arg Arg
}

var _ Simple = Restriction{}

func (Restriction) isSimple() {}

func (r Restriction) Filter() string {
	var sb strings.Builder
	r.WriteFilter(&sb)
	return sb.String()
}

func (r Restriction) WriteFilter(sb *strings.Builder) {
	r.Comparable.WriteFilter(sb)
	if r.Comparator != "" {
		_ = sb.WriteByte(' ')
		r.Comparator.WriteFilter(sb)
		_ = sb.WriteByte(' ')
		r.Arg.WriteFilter(sb)
	}
}

// Comparable may either be a member or function.
//
// EBNF
//
//  comparable
//     : member
//     | function
//     ;
type Comparable interface {
	isComparable()
	isArg()
	WriteFilter(*strings.Builder)
}

var (
	_ Comparable = Member{}
	_ Comparable = Function{}
)

// Member expressions are either value or DOT qualified field references.
//
// Example: `expr.type_map.1.type`
//
// EBNF
//
//  member
//    : value {DOT field}
//    ;
type Member struct {
	// Value of the member.
	Value Value
	// Field of the member.
	Fields []Field
}

var (
	_ Comparable = Member{}
	_ Arg        = Member{}
)

func (Member) isComparable() {}
func (Member) isArg()        {}

func (m Member) Filter() string {
	var sb strings.Builder
	m.WriteFilter(&sb)
	return sb.String()
}

func (m Member) WriteFilter(sb *strings.Builder) {
	m.Value.WriteFilter(sb)
	for _, field := range m.Fields {
		_, _ = sb.WriteString(string(TokenTypeDot))
		field.WriteFilter(sb)
	}
}

// Function calls may use simple or qualified names with zero or more
// arguments.
//
// All functions declared within the list filter, apart from the special
// `arguments` function must be provided by the host service.
//
// Examples:
// * `regex(m.key, '^.*prod.*$')`
// * `math.mem('30mb')`
//
// Antipattern: simple and qualified function names may include keywords:
// NOT, AND, OR. It is not recommended that any of these names be used
// within functions exposed by a service that supports list filters.
//
// EBNF
//
//  function
//    : name {DOT name} LPAREN [argList] RPAREN
//    ;
type Function struct {
	// Names of the function.
	Names []Name
	// Args of the function.
	Args []Arg
}

var _ Comparable = Function{}

func (Function) isComparable() {}
func (Function) isArg()        {}

func (f Function) Filter() string {
	var sb strings.Builder
	f.WriteFilter(&sb)
	return sb.String()
}

func (f Function) WriteFilter(sb *strings.Builder) {
	for i, name := range f.Names {
		name.WriteFilter(sb)
		if i < len(f.Names)-1 {
			_, _ = sb.WriteString(string(TokenTypeDot))
		}
	}
	_, _ = sb.WriteString(string(TokenTypeLeftParen))
	for i, arg := range f.Args {
		arg.WriteFilter(sb)
		if i < len(f.Args)-1 {
			_, _ = sb.WriteString(string(TokenTypeComma))
			_ = sb.WriteByte(' ')
		}
	}
	_, _ = sb.WriteString(string(TokenTypeRightParen))
}

// Comparator in a filter expression.
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
type Comparator string

const (
	ComparatorLessEquals    Comparator = "<="
	ComparatorLessThan      Comparator = "<"
	ComparatorGreaterEquals Comparator = ">="
	ComparatorGreaterThan   Comparator = ">"
	ComparatorNotEquals     Comparator = "!="
	ComparatorEquals        Comparator = "="
	ComparatorHas           Comparator = ":"
)

func (c Comparator) Filter() string {
	var sb strings.Builder
	c.WriteFilter(&sb)
	return sb.String()
}

func (c Comparator) WriteFilter(sb *strings.Builder) {
	_, _ = sb.WriteString(string(c))
}

// Composite is a parenthesized expression, commonly used to group
// terms or clarify operator precedence.
//
// Example: `(msg.endsWith('world') AND retries < 10)`
//
// EBNF
//
//  composite
//    : LPAREN expression RPAREN
//    ;
type Composite struct {
	// Expression of the composite.
	Expression Expression
}

var _ Simple = Composite{}

func (Composite) isArg()    {}
func (Composite) isSimple() {}

func (c Composite) Filter() string {
	var sb strings.Builder
	c.WriteFilter(&sb)
	return sb.String()
}

func (c Composite) WriteFilter(sb *strings.Builder) {
	_, _ = sb.WriteString(string(TokenTypeLeftParen))
	c.Expression.WriteFilter(sb)
	_, _ = sb.WriteString(string(TokenTypeRightParen))
}

// Value may either be a TEXT or STRING.
//
// TEXT is a free-form set of characters without whitespace (WS)
// or . (DOT) within it. The text may represent a variable, string,
// number, boolean, or alternative literal value and must be handled
// in a manner consistent with the service's intention.
//
// STRING is a quoted string which may or may not contain a special
// wildcard `*` character at the beginning or end of the string to
// indicate a prefix or suffix-based search within a restriction.
//
// EBNF
//
// value
//    : TEXT
//    | STRING
//    ;
type Value interface {
	isValue()
	Field
	WriteFilter(*strings.Builder)
}

var (
	_ Value = Text("")
	_ Value = String("")
)

// Field may be either a value or a keyword.
//
// EBNF
//
//  field
//    : value
//    | keyword
//    ;
type Field interface {
	isField()
	WriteFilter(*strings.Builder)
}

var (
	_ Field = Value(nil)
	_ Field = Keyword("")
)

// Name may either be TEXT or a keyword.
//
// EBNF
//
//  name
//    : TEXT
//    | keyword
//    ;
type Name interface {
	isName()
	WriteFilter(*strings.Builder)
}

var (
	_ Name = Text("")
	_ Name = Keyword("")
)

// Arg in a filter expression.
//
// EBNF
//
//  arg
//    : comparable
//    | composite
//    ;
type Arg interface {
	isArg()
	WriteFilter(*strings.Builder)
}

var (
	_ Arg = Comparable(nil)
	_ Arg = Composite{}
)

// Keyword in a filter expression.
//
// EBNF
//
//  keyword
//    : NOT
//    | AND
//    | OR
//    ;
type Keyword string

func (Keyword) isField() {}
func (Keyword) isName()  {}

const (
	KeywordNot Keyword = "NOT"
	KeywordAnd Keyword = "AND"
	KeywordOr  Keyword = "OR"
)

func (k Keyword) WriteFilter(sb *strings.Builder) {
	_, _ = sb.WriteString(string(k))
}

// Text value a filter expression.
type Text string

var (
	_ Value = Text("")
	_ Field = Text("")
	_ Name  = Text("")
)

func (Text) isValue() {}
func (Text) isField() {}
func (Text) isName()  {}

func (t Text) WriteFilter(sb *strings.Builder) {
	_, _ = sb.WriteString(string(t))
}

// String value in a filter expression.
type String string

func (String) isValue() {}
func (String) isField() {}

func (s String) WriteFilter(sb *strings.Builder) {
	_, _ = sb.WriteString(string(s))
}
