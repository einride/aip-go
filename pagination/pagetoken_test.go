package pagination

import (
	"testing"

	freightv1 "go.einride.tech/aip/proto/gen/einride/example/freight/v1"
	"google.golang.org/genproto/googleapis/example/library/v1"
	"gotest.tools/v3/assert"
)

func TestParseOffsetPageToken(t *testing.T) {
	t.Parallel()
	t.Run("valid checksums", func(t *testing.T) {
		t.Parallel()
		request1 := &library.ListBooksRequest{
			Parent:   "shelves/1",
			PageSize: 10,
		}
		pageToken1, err := ParsePageToken(request1)
		assert.NilError(t, err)
		request2 := &library.ListBooksRequest{
			Parent:    "shelves/1",
			PageSize:  20,
			PageToken: pageToken1.Next(request1).String(),
		}
		pageToken2, err := ParsePageToken(request2)
		assert.NilError(t, err)
		assert.Equal(t, int64(10), pageToken2.Offset)
		request3 := &library.ListBooksRequest{
			Parent:    "shelves/1",
			PageSize:  30,
			PageToken: pageToken2.Next(request2).String(),
		}
		pageToken3, err := ParsePageToken(request3)
		assert.NilError(t, err)
		assert.Equal(t, int64(30), pageToken3.Offset)
	})
	t.Run("skip", func(t *testing.T) {
		t.Run("docs example 1", func(t *testing.T) {
			// From https://google.aip.dev/158:
			// A request with no page token and a skip value of 30 returns a single
			// page of results starting with the 31st result.
			pageToken, err := ParsePageToken(&freightv1.ListSitesRequest{
				Parent: "shippers/1",
				Skip:   30,
			})
			assert.NilError(t, err)
			assert.Equal(t, int64(30), pageToken.Offset) // 31st result
		})

		t.Run("docs example 2", func(t *testing.T) {
			// From https://google.aip.dev/158:
			// A request with a page token corresponding to the 51st result (because the first
			// 50 results were returned on the first page) and a skip value of 30 returns a
			// single page of results starting with the 81st result.
			request1 := &freightv1.ListSitesRequest{
				Parent:   "shippers/1",
				PageSize: 50,
			}
			pageToken1, err := ParsePageToken(request1)
			assert.NilError(t, err)
			request2 := &freightv1.ListSitesRequest{
				Parent:    "shippers/1",
				Skip:      30,
				PageSize:  50,
				PageToken: pageToken1.Next(request1).String(),
			}
			pageToken2, err := ParsePageToken(request2)
			assert.NilError(t, err)
			assert.Equal(t, int64(80), pageToken2.Offset)
		})

		t.Run("handle empty token with skip", func(t *testing.T) {
			request1 := &freightv1.ListSitesRequest{
				Parent:   "shippers/1",
				Skip:     30,
				PageSize: 20,
			}
			pageToken1, err := ParsePageToken(request1)
			assert.NilError(t, err)
			assert.Equal(t, int64(30), pageToken1.Offset)
		})
		t.Run("handle existing token with another skip", func(t *testing.T) {
			request1 := &freightv1.ListSitesRequest{
				Parent:   "shippers/1",
				Skip:     50,
				PageSize: 20,
			}
			pageToken1, err := ParsePageToken(request1)
			assert.NilError(t, err)
			assert.Equal(t, int64(50), pageToken1.Offset)
			request2 := &freightv1.ListSitesRequest{
				Parent:    "shippers/1",
				Skip:      30,
				PageSize:  0,
				PageToken: pageToken1.String(),
			}
			pageToken2, err := ParsePageToken(request2)
			assert.NilError(t, err)
			pageToken3 := pageToken2.Next(request2)
			assert.Equal(t, int64(80), pageToken3.Offset)
		})
		t.Run("handle existing token with pagesize and skip", func(t *testing.T) {
			request1 := &freightv1.ListSitesRequest{
				Parent:   "shippers/1",
				Skip:     50,
				PageSize: 20,
			}
			pageToken1, err := ParsePageToken(request1)
			assert.NilError(t, err)
			assert.Equal(t, int64(50), pageToken1.Offset)
			request2 := &freightv1.ListSitesRequest{
				Parent:    "shippers/1",
				Skip:      30,
				PageSize:  20,
				PageToken: pageToken1.String(),
			}
			pageToken2, err := ParsePageToken(request2)
			assert.NilError(t, err)
			pageToken3 := pageToken2.Next(request2)
			assert.Equal(t, int64(100), pageToken3.Offset)
		})
	})
	t.Run("cursor", func(t *testing.T) {
		t.Parallel()
		t.Run("round-trip preserves cursor", func(t *testing.T) {
			t.Parallel()
			request1 := &library.ListBooksRequest{
				Parent:   "shelves/1",
				PageSize: 10,
			}
			pageToken1, err := ParsePageToken(request1)
			assert.NilError(t, err)
			pageToken1.Cursor = []any{"abc", int64(42)}
			request2 := &library.ListBooksRequest{
				Parent:    "shelves/1",
				PageSize:  10,
				PageToken: pageToken1.String(),
			}
			pageToken2, err := ParsePageToken(request2)
			assert.NilError(t, err)
			assert.DeepEqual(t, []any{"abc", int64(42)}, pageToken2.Cursor)
		})
		t.Run("Next preserves cursor", func(t *testing.T) {
			t.Parallel()
			request := &library.ListBooksRequest{
				Parent:   "shelves/1",
				PageSize: 10,
			}
			pageToken, err := ParsePageToken(request)
			assert.NilError(t, err)
			pageToken.Cursor = []any{"abc", int64(42)}
			next := pageToken.Next(request)
			assert.Equal(t, int64(10), next.Offset)
			assert.DeepEqual(t, []any{"abc", int64(42)}, next.Cursor)
		})
		t.Run("empty cursor round-trips as nil", func(t *testing.T) {
			t.Parallel()
			request1 := &library.ListBooksRequest{
				Parent:   "shelves/1",
				PageSize: 10,
			}
			pageToken1, err := ParsePageToken(request1)
			assert.NilError(t, err)
			request2 := &library.ListBooksRequest{
				Parent:    "shelves/1",
				PageSize:  10,
				PageToken: pageToken1.String(),
			}
			pageToken2, err := ParsePageToken(request2)
			assert.NilError(t, err)
			assert.Assert(t, pageToken2.Cursor == nil)
		})
		t.Run("token encoded before cursor field is added", func(t *testing.T) {
			t.Parallel()
			// Token shape from before the Cursor field was added.
			type legacyPageToken struct {
				Offset          int64
				RequestChecksum uint32
			}
			request := &library.ListBooksRequest{
				Parent:   "shelves/1",
				PageSize: 10,
			}
			checksum, err := CalculateRequestChecksum(request)
			assert.NilError(t, err)
			checksum ^= pageTokenChecksumMask
			legacy := EncodePageTokenStruct(&legacyPageToken{
				Offset:          42,
				RequestChecksum: checksum,
			})
			parsed, err := ParsePageToken(&library.ListBooksRequest{
				Parent:    "shelves/1",
				PageSize:  10,
				PageToken: legacy,
			})
			assert.NilError(t, err)
			assert.Equal(t, int64(42), parsed.Offset)
			assert.Assert(t, parsed.Cursor == nil)
		})
		t.Run("NextCursor populates cursor from message fields", func(t *testing.T) {
			t.Parallel()
			request := &library.ListBooksRequest{
				Parent:   "shelves/1",
				PageSize: 10,
			}
			pageToken, err := ParsePageToken(request)
			assert.NilError(t, err)
			book := &library.Book{
				Name:   "shelves/1/books/42",
				Author: "Ada",
				Read:   true,
			}
			next, err := pageToken.NextCursor(book, "name", "author", "read")
			assert.NilError(t, err)
			assert.DeepEqual(t, []any{"shelves/1/books/42", "Ada", true}, next.Cursor)
		})
		t.Run("NextCursor round-trips through page token", func(t *testing.T) {
			t.Parallel()
			request1 := &library.ListBooksRequest{
				Parent:   "shelves/1",
				PageSize: 10,
			}
			pageToken1, err := ParsePageToken(request1)
			assert.NilError(t, err)
			pageToken1, err = pageToken1.NextCursor(&library.Book{Name: "shelves/1/books/7"}, "name")
			assert.NilError(t, err)
			request2 := &library.ListBooksRequest{
				Parent:    "shelves/1",
				PageSize:  10,
				PageToken: pageToken1.String(),
			}
			pageToken2, err := ParsePageToken(request2)
			assert.NilError(t, err)
			assert.DeepEqual(t, []any{"shelves/1/books/7"}, pageToken2.Cursor)
		})
		t.Run("NextCursor errors on unknown field", func(t *testing.T) {
			t.Parallel()
			_, err := PageToken{}.NextCursor(&library.Book{}, "nonexistent")
			assert.ErrorContains(t, err, "not found")
		})
	})
	t.Run("invalid format", func(t *testing.T) {
		t.Parallel()
		request := &library.ListBooksRequest{
			Parent:    "shelves/1",
			PageSize:  10,
			PageToken: "invalid",
		}
		pageToken1, err := ParsePageToken(request)
		assert.ErrorContains(t, err, "decode")
		assert.DeepEqual(t, PageToken{}, pageToken1)
	})

	t.Run("invalid checksum", func(t *testing.T) {
		t.Parallel()
		request := &library.ListBooksRequest{
			Parent:   "shelves/1",
			PageSize: 10,
			PageToken: EncodePageTokenStruct(&PageToken{
				Offset:          100,
				RequestChecksum: 1234, // invalid
			}),
		}
		pageToken1, err := ParsePageToken(request)
		assert.ErrorContains(t, err, "checksum")
		assert.DeepEqual(t, PageToken{}, pageToken1)
	})
}
