package resourcename

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestScan(t *testing.T) {
	t.Parallel()
	t.Run("no variables", func(t *testing.T) {
		t.Parallel()
		assert.NilError(
			t,
			Sscan(
				"publishers",
				"publishers",
			),
		)
	})

	t.Run("single variable", func(t *testing.T) {
		t.Parallel()
		var publisher string
		assert.NilError(
			t,
			Sscan(
				"publishers/foo",
				"publishers/{publisher}",
				&publisher,
			),
		)
		assert.Equal(t, "foo", publisher)
	})

	t.Run("two variables", func(t *testing.T) {
		t.Parallel()
		var publisher, book string
		assert.NilError(
			t,
			Sscan(
				"publishers/foo/books/bar",
				"publishers/{publisher}/books/{book}",
				&publisher,
				&book,
			),
		)
		assert.Equal(t, "foo", publisher)
		assert.Equal(t, "bar", book)
	})

	t.Run("two variables singleton", func(t *testing.T) {
		t.Parallel()
		var publisher, book string
		assert.NilError(
			t,
			Sscan(
				"publishers/foo/books/bar/settings",
				"publishers/{publisher}/books/{book}/settings",
				&publisher,
				&book,
			),
		)
		assert.Equal(t, "foo", publisher)
		assert.Equal(t, "bar", book)
	})

	t.Run("two variables singleton", func(t *testing.T) {
		t.Parallel()
		var publisher, book string
		assert.NilError(
			t,
			Sscan(
				"publishers/foo/books/bar/settings",
				"publishers/{publisher}/books/{book}/settings",
				&publisher,
				&book,
			),
		)
		assert.Equal(t, "foo", publisher)
		assert.Equal(t, "bar", book)
	})

	t.Run("trailing segments", func(t *testing.T) {
		t.Parallel()
		var publisher, book string
		assert.ErrorContains(
			t,
			Sscan(
				"publishers/foo/books/bar/settings",
				"publishers/{publisher}/books/{book}",
				&publisher,
				&book,
			),
			"trailing",
		)
	})

	t.Run("too few variables", func(t *testing.T) {
		t.Parallel()
		var publisher string
		assert.ErrorContains(
			t,
			Sscan(
				"publishers/foo/books/bar/settings",
				"publishers/{publisher}/books/{book}",
				&publisher,
			),
			"too few variables",
		)
	})

	t.Run("too many variables", func(t *testing.T) {
		t.Parallel()
		var publisher, book, extra string
		assert.ErrorContains(
			t,
			Sscan(
				"publishers/foo/books/bar",
				"publishers/{publisher}/books/{book}",
				&publisher,
				&book,
				&extra,
			),
			"too many variables",
		)
	})
}

// nolint: gochecknoglobals
var benchmarkScanSink string

func BenchmarkScan(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var publisher, book string
		if err := Sscan(
			"publishers/foo/books/bar",
			"publishers/{publisher}/books/{book}",
			&publisher,
			&book,
		); err != nil {
			b.Fatal(err)
		}
		benchmarkScanSink = publisher
	}
}
