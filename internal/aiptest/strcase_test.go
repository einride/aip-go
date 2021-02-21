package aiptest

import (
	"fmt"
	"testing"
	"unicode"

	"gotest.tools/v3/assert"
)

func Test_camelToSnake(t *testing.T) {
	t.Parallel()
	fmt.Println(unicode.IsUpper('_'))
	for _, tt := range []struct {
		in  string
		out string
	}{
		{
			in:  "single",
			out: "single",
		},
		{
			in:  "multipleWords",
			out: "multiple_words",
		},
		{
			in:  "StartUpper",
			out: "start_upper",
		},
	} {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			assert.DeepEqual(t, tt.out, camelToSnake(tt.in))
		})
	}
}
