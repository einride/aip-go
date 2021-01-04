package filtering

// TokenType represents the type of a filter expression token.
//
// See: https://google.aip.dev/assets/misc/ebnf-filtering.txt
type TokenType string

// Value token types.
const (
	TokenTypeWhitespace TokenType = "WS"
	TokenTypeText       TokenType = "TEXT"
	TokenTypeString     TokenType = "STRING"
)

// Keyword token types.
const (
	TokenTypeNot TokenType = "NOT"
	TokenTypeAnd TokenType = "AND"
	TokenTypeOr  TokenType = "OR"
)

// Operator token types.
const (
	TokenTypeLeftParen     TokenType = "("
	TokenTypeRightParen    TokenType = ")"
	TokenTypeMinus         TokenType = "-"
	TokenTypeDot           TokenType = "."
	TokenTypeEquals        TokenType = "="
	TokenTypeHas           TokenType = ":"
	TokenTypeLess          TokenType = "<"
	TokenTypeGreater       TokenType = ">"
	TokenTypeExclaim       TokenType = "!"
	TokenTypeComma         TokenType = ","
	TokenTypeLessEquals    TokenType = "<="
	TokenTypeGreaterEquals TokenType = ">="
	TokenTypeNotEquals     TokenType = "!="
)
