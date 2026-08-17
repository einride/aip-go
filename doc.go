// Package flop implements declarative AIP-132 ordering, AIP-160 filtering,
// and cursor or page-number pagination. Paginator validates request input and
// exposes backend-neutral parameters to storage adapters.
//
// ParseFilter parses an AIP-160 filter into the expr expression tree, and
// FilterSchema validates a parsed filter against the fields, types, and limits
// one endpoint declares.
package flop
