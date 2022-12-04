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
	t.Run("invalid format", func(t *testing.T) {
		t.Parallel()
		request := &library.ListBooksRequest{
			Parent:    "shelves/1",
			PageSize:  10,
			PageToken: "invalid",
		}
		pageToken1, err := ParsePageToken(request)
		assert.ErrorContains(t, err, "decode")
		assert.Equal(t, PageToken{}, pageToken1)
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
		assert.Equal(t, PageToken{}, pageToken1)
	})
}
