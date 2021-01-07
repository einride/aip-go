package aipreflect

import (
	"unicode"
	"unicode/utf8"
)

func initialUpperCase(s string) string {
	if len(s) == 0 {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}
