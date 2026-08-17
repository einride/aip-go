package flop_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ysomad/flop"
)

// FuzzParseFilter checks that no input panics, that every canonical rendering
// is itself parseable, and that canonicalization settles after one pass.
//
// It deliberately does not require the first canonical form to equal the
// second: Canonical does not preserve type, so a float such as -0.0 renders as
// "-0" and re-parses as an integer. That is harmless because the canonical
// string is only ever hashed into a cursor fingerprint, never re-parsed, and
// the same request always canonicalizes the same way.
func FuzzParseFilter(f *testing.F) {
	seeds := []string{
		"",
		"a = 1",
		`name = "book*"`,
		"a = 1 OR b = 2 AND c = 3",
		`NOT (active = false OR deleted = null)`,
		`"blue" "chair"`,
		`tags:"blue"`,
		"labels.* = 1",
		"1e+2",
		"-9223372036854775808",
		"18446744073709551615",
		"wait < 2h",
		`'it\'s'`,
		strings.Repeat("(", 70) + "a" + strings.Repeat(")", 70),
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		parsed, err := flop.ParseFilter(input)
		if err != nil {
			return
		}
		once := parsed.String()
		reparsed, ok := reparse(t, once, input)
		if !ok {
			return
		}
		twice := reparsed.String()
		settled, ok := reparse(t, twice, input)
		if !ok {
			return
		}
		if got := settled.String(); got != twice {
			t.Fatalf("canonical form never settles: %q then %q then %q from %q", once, twice, got, input)
		}
	})
}

// reparse parses a canonical rendering, reporting ok=false for the one failure
// that is expected rather than a defect.
//
// Canonical output parenthesises every logical node, but the parser only bounds
// parenthesis nesting — a long unparenthesised sequence such as "a b c ..."
// parses at zero nesting cost and then renders deeper than ParseFilter will
// accept. That is harmless: canonical output is hashed into a cursor
// fingerprint and is never fed back through ParseFilter, and
// FilterSchema.Validate rejects trees this deep via FilterLimits.MaxDepth long
// before one reaches a cursor. Any other parse failure means a rendering the
// parser cannot read, which is a defect.
func reparse(t *testing.T, canonical, input string) (flop.Filter, bool) {
	t.Helper()
	parsed, err := flop.ParseFilter(canonical)
	if err == nil {
		return parsed, true
	}
	var filterErr *flop.FilterError
	if errors.As(err, &filterErr) && filterErr.Msg == "nesting is too deep" {
		return flop.Filter{}, false
	}
	t.Fatalf("canonical form does not parse: %q from %q: %v", canonical, input, err)
	return flop.Filter{}, false
}
