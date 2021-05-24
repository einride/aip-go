package resourcename

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestValidate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		input         string
		errorContains string
	}{
		{
			name:          "empty",
			input:         "",
			errorContains: "empty",
		},

		{
			name:          "invalid DNS characters",
			input:         "ice cream is best",
			errorContains: "not a valid DNS name",
		},

		{
			name:          "invalid DNS characters in segment",
			input:         "foo/bar/ice cream is best",
			errorContains: "not a valid DNS name",
		},

		{
			name:          "invalid DNS characters in domain",
			input:         "//ice cream is best.com/foo/bar",
			errorContains: "not a valid DNS name",
		},

		{
			name:  "singleton",
			input: "foo",
		},

		{
			name:  "singleton wildcard",
			input: "-",
		},

		{
			name:  "multi",
			input: "foo/bar",
		},

		{
			name:  "multi wildcard at start",
			input: "-/bar",
		},

		{
			name:  "multi wildcard at end",
			input: "foo/-",
		},

		{
			name:  "multi wildcard at middle",
			input: "foo/-/bar",
		},

		{
			name:  "numeric",
			input: "foo/1234/bar",
		},

		{
			name:  "camelCase",
			input: "FOO/1234/bAr",
		},

		{
			name:  "full",
			input: "//example.com/foo/bar",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(tt.input)
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
