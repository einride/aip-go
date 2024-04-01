package single

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestBookResourceName(t *testing.T) {
	t.Run("good", func(t *testing.T) {
		const (
			shelf = "shelf"
			book  = "book"
		)
		pattern := fmt.Sprintf("shelves/%s/books/%s", shelf, book)
		var name BookResourceName
		err := name.UnmarshalString(pattern)
		assert.NilError(t, err)
		assert.Equal(t, name.Shelf, shelf)
		assert.Equal(t, name.Book, book)

		marshalled, err := name.MarshalString()
		assert.NilError(t, err)
		assert.Equal(t, marshalled, pattern)
	})

	t.Run("bad top-level", func(t *testing.T) {
		var name BookResourceName
		err := name.UnmarshalString("books/book")
		assert.Error(
			t,
			err,
			"parse resource name 'books/book' with pattern 'shelves/{shelf}/books/{book}': segment shelves: got books",
		)
	})

	t.Run("bad wrong parent", func(t *testing.T) {
		var name BookResourceName
		err := name.UnmarshalString("others/other/books/book")
		assert.Error(
			t,
			err,
			"parse resource name 'others/other/books/book' with pattern 'shelves/{shelf}/books/{book}': segment shelves: got others",
		)
	})
}

func TestShelfResourceName(t *testing.T) {
	t.Run("good", func(t *testing.T) {
		const shelf = "shelf"
		pattern := fmt.Sprintf("shelves/%s", shelf)
		var name ShelfResourceName
		err := name.UnmarshalString(pattern)
		assert.NilError(t, err)
		assert.Equal(t, name.Shelf, shelf)

		marshalled, err := name.MarshalString()
		assert.NilError(t, err)
		assert.Equal(t, marshalled, pattern)
	})

	t.Run("invalid", func(t *testing.T) {
		var name ShelfResourceName
		err := name.UnmarshalString("others/other")
		assert.Error(t, err, "parse resource name 'others/other' with pattern 'shelves/{shelf}': segment shelves: got others")
	})

	t.Run("bad wrong parent", func(t *testing.T) {
		var name ShelfResourceName
		err := name.UnmarshalString("others/other/shelves/shelf")
		assert.Error(
			t,
			err,
			"parse resource name 'others/other/shelves/shelf' with pattern 'shelves/{shelf}': segment shelves: got others",
		)
	})
}
