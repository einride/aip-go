package originallysinglepattern

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestParseBookMultiPatternResourceName(t *testing.T) {
	goodPatterns := []string{
		"shelves/shelf/books/book",
		"publishers/publisher/books/book",
	}
	for _, pattern := range goodPatterns {
		pattern := pattern
		t.Run(pattern, func(t *testing.T) {
			name, err := ParseBookMultiPatternResourceName(pattern)
			assert.NilError(t, err)
			assert.Equal(t, name.String(), pattern)
		})
	}

	badPatterns := []string{
		"books/book",
		"others/other",
		"others/other/books/book",
	}
	for _, pattern := range badPatterns {
		pattern := pattern
		t.Run(pattern, func(t *testing.T) {
			_, err := ParseBookMultiPatternResourceName(pattern)
			assert.Error(t, err, "no matching pattern")
		})
	}
}

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
		assert.Error(t, err, "parse resource name 'books/book' with pattern 'shelves/{shelf}/books/{book}': segment shelves: got books")
	})

	t.Run("bad wrong parent", func(t *testing.T) {
		var name BookResourceName
		err := name.UnmarshalString("others/other/books/book")
		assert.Error(t, err, "parse resource name 'others/other/books/book' with pattern 'shelves/{shelf}/books/{book}': segment shelves: got others")
	})

	t.Run("bad newer parent", func(t *testing.T) {
		var name BookResourceName
		err := name.UnmarshalString("publishers/publisher/books/book")
		assert.Error(t, err, "parse resource name 'publishers/publisher/books/book' with pattern 'shelves/{shelf}/books/{book}': segment shelves: got publishers")
	})
}

func TestShelvesBookResourceName(t *testing.T) {
	t.Run("good", func(t *testing.T) {
		const (
			shelf = "shelf"
			book  = "book"
		)
		pattern := fmt.Sprintf("shelves/%s/books/%s", shelf, book)
		var name ShelvesBookResourceName
		err := name.UnmarshalString(pattern)
		assert.NilError(t, err)
		assert.Equal(t, name.Shelf, shelf)
		assert.Equal(t, name.Book, book)

		marshalled, err := name.MarshalString()
		assert.NilError(t, err)
		assert.Equal(t, marshalled, pattern)
	})

	t.Run("bad top-level", func(t *testing.T) {
		var name ShelvesBookResourceName
		err := name.UnmarshalString("books/book")
		assert.Error(t, err, "parse resource name 'books/book' with pattern 'shelves/{shelf}/books/{book}': segment shelves: got books")
	})

	t.Run("bad wrong parent", func(t *testing.T) {
		var name ShelvesBookResourceName
		err := name.UnmarshalString("others/other/books/book")
		assert.Error(t, err, "parse resource name 'others/other/books/book' with pattern 'shelves/{shelf}/books/{book}': segment shelves: got others")
	})
}

func TestPublishersBookResourceName(t *testing.T) {
	t.Run("good", func(t *testing.T) {
		const (
			publisher = "publisher"
			book      = "book"
		)
		pattern := fmt.Sprintf("publishers/%s/books/%s", publisher, book)
		var name PublishersBookResourceName
		err := name.UnmarshalString(pattern)
		assert.NilError(t, err)
		assert.Equal(t, name.Publisher, publisher)
		assert.Equal(t, name.Book, book)

		marshalled, err := name.MarshalString()
		assert.NilError(t, err)
		assert.Equal(t, marshalled, pattern)
	})

	t.Run("bad top-level", func(t *testing.T) {
		var name PublishersBookResourceName
		err := name.UnmarshalString("books/book")
		assert.Error(t, err, "parse resource name 'books/book' with pattern 'publishers/{publisher}/books/{book}': segment publishers: got books")
	})

	t.Run("bad wrong parent", func(t *testing.T) {
		var name PublishersBookResourceName
		err := name.UnmarshalString("others/other/books/book")
		assert.Error(t, err, "parse resource name 'others/other/books/book' with pattern 'publishers/{publisher}/books/{book}': segment publishers: got others")
	})
}
