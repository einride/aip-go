package validation

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"
)

func TestMessageValidator(t *testing.T) {
	t.Parallel()

	t.Run("no violation", func(t *testing.T) {
		t.Parallel()
		var v MessageValidator
		assert.NilError(t, v.Err())
	})

	t.Run("add single violation", func(t *testing.T) {
		t.Parallel()
		var v MessageValidator
		v.AddFieldViolation("foo", "bar")
		assert.Error(t, v.Err(), "field violation on foo: bar")
	})

	t.Run("add single violation with parent", func(t *testing.T) {
		t.Parallel()
		var v MessageValidator
		v.SetParentField("foo")
		v.AddFieldViolation("bar", "baz")
		assert.Error(t, v.Err(), "field violation on foo.bar: baz")
	})

	t.Run("add nested violations", func(t *testing.T) {
		t.Parallel()
		var inner MessageValidator
		inner.AddFieldViolation("b", "c")
		var outer MessageValidator
		outer.AddFieldError("a", inner.Err())
		assert.Error(t, outer.Err(), "field violation on a.b: c")
	})

	t.Run("add field error", func(t *testing.T) {
		t.Parallel()
		var v MessageValidator
		v.AddFieldError("a", errors.New("boom"))
		assert.Error(t, v.Err(), "field violation on a: boom")
	})
}
