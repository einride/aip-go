// Package expr defines the filter expression vocabulary: the typed value
// literals and the expression tree an AIP-160 filter parses into.
//
// It holds no parsing or validation logic. The filter package produces these
// types; the flop package and storage adapters consume them.
package expr
