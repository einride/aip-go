package pagination

import (
	"testing"

	"google.golang.org/genproto/googleapis/example/library/v1"
	"gotest.tools/v3/assert"
)

func TestParseOffsetPageToken(t *testing.T) {
	t.Parallel()
	t.Run("valid checksums", func(t *testing.T) {
		t.Parallel()
		request1 := &library.ListBooksRequest{
			Name:     "shelves/1",
			PageSize: 10,
		}
		pageToken1, err := ParsePageToken(request1)
		assert.NilError(t, err)
		request2 := &library.ListBooksRequest{
			Name:      "shelves/1",
			PageSize:  20,
			PageToken: pageToken1.Next(request1).String(),
		}
		pageToken2, err := ParsePageToken(request2)
		assert.NilError(t, err)
		assert.Equal(t, int64(10), pageToken2.Offset)
		request3 := &library.ListBooksRequest{
			Name:      "shelves/1",
			PageSize:  30,
			PageToken: pageToken2.Next(request2).String(),
		}
		pageToken3, err := ParsePageToken(request3)
		assert.NilError(t, err)
		assert.Equal(t, int64(30), pageToken3.Offset)
	})

	t.Run("invalid format", func(t *testing.T) {
		t.Parallel()
		request := &library.ListBooksRequest{
			Name:      "shelves/1",
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
			Name:     "shelves/1",
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
