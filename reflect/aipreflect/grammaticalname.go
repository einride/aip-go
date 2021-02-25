package aipreflect

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// GrammaticalName is the grammatical name for the singular or plural form of resource type.
// Grammatical names must be URL-safe and use lowerCamelCase.
type GrammaticalName string // e.g. "userEvents"

// Validate checks that the grammatical name is non-empty, URL-safe, and uses lowerCamelCase.
func (g GrammaticalName) Validate() error {
	if len(g) == 0 {
		return fmt.Errorf("validate grammatical name: must be non-empty")
	}
	for _, r := range g {
		if !unicode.In(r, unicode.Letter, unicode.Digit) {
			return fmt.Errorf("validate grammatical name '%s': contains forbidden character '%s'", g, string(r))
		}
	}
	if r, _ := utf8.DecodeRuneInString(string(g)); !unicode.IsLower(r) {
		return fmt.Errorf("validate grammatical name '%s': must be lowerCamelCase", g)
	}
	return nil
}

// UpperCamelCase returns the UpperCamelCase version of the grammatical name, for use in e.g. method names.
func (g GrammaticalName) UpperCamelCase() string {
	return initialUpperCase(string(g))
}
