package ident

import (
	"testing"
	"unicode/utf8"

	"github.com/ysomad/flop/internal/assert"
)

func TestIsStart(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input rune
		want  bool
	}{
		{name: "underscore", input: '_', want: true},
		{name: "upper first", input: 'A', want: true},
		{name: "upper last", input: 'Z', want: true},
		{name: "lower first", input: 'a', want: true},
		{name: "lower last", input: 'z', want: true},
		{name: "before upper range", input: '@'},
		{name: "after upper range", input: '['},
		{name: "before lower range", input: '`'},
		{name: "after lower range", input: '{'},
		{name: "digit first", input: '0'},
		{name: "digit last", input: '9'},
		{name: "dot", input: '.'},
		{name: "hyphen", input: '-'},
		{name: "space", input: ' '},
		{name: "cyrillic", input: 'и'},
		{name: "latin with diacritic", input: 'É'},
		{name: "replacement rune", input: utf8.RuneError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, IsStart(test.input))
		})
	}
}

func TestIsContinue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input rune
		want  bool
	}{
		{name: "underscore", input: '_', want: true},
		{name: "upper first", input: 'A', want: true},
		{name: "upper last", input: 'Z', want: true},
		{name: "lower first", input: 'a', want: true},
		{name: "lower last", input: 'z', want: true},
		{name: "digit first", input: '0', want: true},
		{name: "digit middle", input: '5', want: true},
		{name: "digit last", input: '9', want: true},
		{name: "before digit range", input: '/'},
		{name: "after digit range", input: ':'},
		{name: "dot", input: '.'},
		{name: "hyphen", input: '-'},
		{name: "space", input: ' '},
		{name: "cyrillic", input: 'и'},
		{name: "latin with diacritic", input: 'É'},
		{name: "replacement rune", input: utf8.RuneError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, IsContinue(test.input))
		})
	}
}

func TestValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "single letter", input: "a", want: true},
		{name: "single underscore", input: "_", want: true},
		{name: "leading underscore", input: "_x", want: true},
		{name: "letter and digit", input: "A1", want: true},
		{name: "snake case", input: "user_id", want: true},
		{name: "mixed", input: "Z9_z", want: true},
		{name: "empty"},
		{name: "starts with digit", input: "1a"},
		{name: "dotted path", input: "user.name"},
		{name: "hyphen", input: "user-id"},
		{name: "non-ascii", input: "имя"},
		{name: "inner space", input: "a b"},
		{name: "trailing space", input: "id "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, Valid(test.input))
		})
	}
}

func TestValidPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "single identifier", input: "id", want: true},
		{name: "two parts", input: "user.name", want: true},
		{name: "three parts", input: "a.b.c", want: true},
		{name: "underscored parts", input: "_a._b1", want: true},
		{name: "empty"},
		{name: "dot only", input: "."},
		{name: "leading dot", input: ".id"},
		{name: "trailing dot", input: "id."},
		{name: "empty part", input: "user..name"},
		{name: "part starts with digit", input: "user.1st"},
		{name: "space in part", input: "user name.id"},
		{name: "non-ascii part", input: "имя.id"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, ValidPath(test.input))
		})
	}
}
