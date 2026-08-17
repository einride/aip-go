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
func Valid(id string) bool {
	runes := []rune(id)
	if len(runes) == 0 || !IsStart(runes[0]) {
		return false
	}
	for _, r := range runes[1:] {
		if !IsContinue(r) {
			return false
		}
	}
	return true
}

// ValidPath reports whether path is a dot-separated sequence of identifiers.
func ValidPath(path string) bool {
	if path == "" {
		return false
	}
	for _, part := range strings.Split(path, ".") {
		if !Valid(part) {
			return false
		}
	}
	return true
}
