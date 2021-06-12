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
			name:          "variable",
			input:         "foo/bar/{baz}",
			errorContains: "must not contain variables",
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

func TestValidatePattern(t *testing.T) {
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
			errorContains: "patterns can not be full resource names",
		},

		{
			name:  "variable",
			input: "foo/bar/{baz}",
		},

		{
			name:  "singleton",
			input: "foo",
		},

		{
			name:          "singleton wildcard",
			input:         "-",
			errorContains: "wildcards not allowed in patterns",
		},

		{
			name:  "multi",
			input: "foo/bar",
		},

		{
			name:          "multi wildcard at start",
			input:         "-/bar",
			errorContains: "wildcards not allowed in patterns",
		},

		{
			name:          "multi wildcard at end",
			input:         "foo/-",
			errorContains: "wildcards not allowed in patterns",
		},

		{
			name:          "multi wildcard at middle",
			input:         "foo/-/bar",
			errorContains: "wildcards not allowed in patterns",
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
			name:          "full",
			input:         "//example.com/foo/bar",
			errorContains: "patterns can not be full resource names",
		},

		{
			name:          "invalid variable name",
			input:         "fooBars/{fooBar}",
			errorContains: "must be valid snake case",
		},

		{
			name:  "valid variable name",
			input: "fooBars/{foo_bar}",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidatePattern(tt.input)
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
