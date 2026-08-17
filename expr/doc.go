// Package expr defines the filter expression vocabulary: the typed value
// literals and the expression tree an AIP-160 filter parses into.
//
// It holds no parsing or validation logic. The flop package produces these
// types; flop itself and storage adapters consume them.
package expr
