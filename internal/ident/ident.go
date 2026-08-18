// Package ident implements the ASCII identifier rule shared by order_by
// paths, filter field paths, and cursor position paths.
package ident

import "strings"

// IsStart reports whether r may begin an identifier.
func IsStart(r rune) bool {
	return r == '_' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z'
}

// IsContinue reports whether r may appear after the first rune of an identifier.
func IsContinue(r rune) bool {
	return IsStart(r) || r >= '0' && r <= '9'
}

// Valid reports whether id is a single well-formed identifier.
//
// The rule is ASCII-only, so id is scanned as bytes. Any multi-byte rune has
// every byte >= 0x80 and fails IsContinue, which rejects it just as a rune
// scan would, without decoding.
func Valid(id string) bool {
	if id == "" || !IsStart(rune(id[0])) {
		return false
	}
	for i := 1; i < len(id); i++ {
		if !IsContinue(rune(id[i])) {
			return false
		}
	}
	return true
}

// ValidPath reports whether path is a dot-separated sequence of identifiers.
//
// Segments are walked in place rather than split out, so validating a path
// does not allocate.
func ValidPath(path string) bool {
	if path == "" {
		return false
	}
	for {
		dot := strings.IndexByte(path, '.')
		if dot < 0 {
			return Valid(path)
		}
		if !Valid(path[:dot]) {
			return false
		}
		path = path[dot+1:]
	}
}
