package filtering

import (
	"errors"
	"strings"
	"testing"
)

// The following code is adapted from
// https://github.com/google/cel-go/blob/f6d3c92171c2c8732a8d0a4b24d6729df4261520/parser/unescape.go#L1-L237.

func TestUnescape(t *testing.T) {
	tests := []struct {
		in  string
		out interface{}
	}{
		// Simple string unescaping tests.
		{in: `'hello'`, out: `hello`},
		{in: `""`, out: ``},
		{in: `"\\\""`, out: `\"`},
		{in: `"\\"`, out: `\`},
		{in: `"\303\277"`, out: `Ã¿`},
		{in: `"\377"`, out: `ÿ`},
		{in: `"\u263A\u263A"`, out: `☺☺`},
		{in: `"\a\b\f\n\r\t\v\'\"\\\? Legal escapes"`, out: "\a\b\f\n\r\t\v'\"\\? Legal escapes"},
		// Escaping errors.
		{in: `"\a\b\f\n\r\t\v\'\"\\\? Illegal escape \>"`, out: errors.New("unable to unescape string")},
		{in: `"\u00f"`, out: errors.New("unable to unescape string")},
		{in: `"\u00fÿ"`, out: errors.New("unable to unescape string")},
		{in: `"\26"`, out: errors.New("unable to unescape octal sequence")},
		{in: `"\268"`, out: errors.New("unable to unescape octal sequence")},
		{in: `"\267\"`, out: errors.New(`found '\' as last character`)},
		{in: `'`, out: errors.New("unable to unescape string")},
		{in: `*hello*`, out: errors.New("unable to unescape string")},
	}

	for _, tst := range tests {
		tc := tst
		t.Run(tc.in, func(t *testing.T) {
			got, err := unescape(tc.in)
			if err != nil {
				expect, isErr := tc.out.(error)
				if isErr {
					if !strings.Contains(err.Error(), expect.Error()) {
						t.Errorf("unescape(%s) errored with %v, wanted %v", tc.in, err, expect)
					}
				} else {
					t.Fatalf("unescape(%s) failed: %v", tc.in, err)
				}
			} else if got != tc.out {
				t.Errorf("unescape(%s) got %v, wanted %v", tc.in, got, tc.out)
			}
		})
	}
}
