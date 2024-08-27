package pagination

import (
	"testing"

	freightv1 "go.einride.tech/aip/proto/gen/einride/example/freight/v1"
	"google.golang.org/genproto/googleapis/example/library/v1"
	"gotest.tools/v3/assert"
)

func TestParseCursorPageToken(t *testing.T) {
	t.Parallel()
	t.Run("valid checksums", func(t *testing.T) {
		t.Parallel()
		request1 := &library.ListBooksRequest{
			Parent:   "shelves/1",
			PageSize: 10,
		}
		pageToken1, err := ParseCursor[int](request1)
		assert.NilError(t, err)
		assert.Assert(t, pageToken1.Cursor == nil)
		request2 := &library.ListBooksRequest{
			Parent:    "shelves/1",
			PageSize:  20,
			PageToken: pageToken1.Next(21).String(),
		}
		pageToken2, err := ParseCursor[int](request2)
		assert.NilError(t, err)
		assert.Equal(t, 21, *pageToken2.Cursor)
		request3 := &library.ListBooksRequest{
			Parent:    "shelves/1",
			PageSize:  30,
			PageToken: pageToken2.Next(51).String(),
		}
		pageToken3, err := ParseCursor[int](request3)
		assert.NilError(t, err)
		assert.Equal(t, 51, *pageToken3.Cursor)
	})
	t.Run("skip", func(t *testing.T) {
		t.Run("handle empty token with skip", func(t *testing.T) {
			request1 := &freightv1.ListSitesRequest{
				Parent:   "shippers/1",
				Skip:     30,
				PageSize: 20,
			}
			pageToken1, err := ParseCursor[int](request1)
			assert.NilError(t, err)
			assert.Equal(t, int64(30), pageToken1.Skip)
			assert.Assert(t, pageToken1.Cursor == nil)
		})
		t.Run("handle existing token with another skip", func(t *testing.T) {
			request1 := &freightv1.ListSitesRequest{
				Parent:   "shippers/1",
				Skip:     50,
				PageSize: 20,
			}
			pageToken1, err := ParseCursor[int](request1)
			assert.NilError(t, err)
			assert.Equal(t, int64(50), pageToken1.Skip)
			request2 := &freightv1.ListSitesRequest{
				Parent:    "shippers/1",
				Skip:      30,
				PageToken: pageToken1.String(),
			}
			pageToken2, err := ParseCursor[int](request2)
			assert.Equal(t, int64(30), pageToken2.Skip)
			assert.NilError(t, err)
			pageToken3 := pageToken2.Next(31)
			assert.Equal(t, int64(0), pageToken3.Skip)
		})
	})
	t.Run("invalid format", func(t *testing.T) {
		t.Parallel()
		request := &library.ListBooksRequest{
			Parent:    "shelves/1",
			PageSize:  10,
			PageToken: "invalid",
		}
		pageToken1, err := ParseCursor[int](request)
		assert.ErrorContains(t, err, "decode")
		assert.Equal(t, CursorToken[int]{}, pageToken1)
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
		pageToken1, err := ParseCursor[int](request)
		assert.ErrorContains(t, err, "checksum")
		assert.Equal(t, CursorToken[int]{}, pageToken1)
	})
	t.Run("invalid cursor type", func(t *testing.T) {
		t.Parallel()
		request1 := &library.ListBooksRequest{
			Parent:   "shelves/1",
			PageSize: 10,
		}
		pageToken1, err := ParseCursor[int](request1)
		assert.NilError(t, err)
		assert.Assert(t, pageToken1.Cursor == nil)
		request2 := &library.ListBooksRequest{
			Parent:    "shelves/1",
			PageSize:  20,
			PageToken: pageToken1.Next(21).String(),
		}
		pageToken2, err := ParseCursor[string](request2)
		assert.ErrorContains(t, err, "gob: wrong type")
		assert.Equal(t, CursorToken[string]{}, pageToken2)
	})
}
