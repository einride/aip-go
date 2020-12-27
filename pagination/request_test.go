package pagination

import (
	"testing"

	"google.golang.org/genproto/googleapis/example/library/v1"
	"gotest.tools/v3/assert"
)

func TestCalculateRequestChecksum(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		request1 Request
		request2 Request
		equal    bool
	}{
		{
			name: "same request",
			request1: &library.ListBooksRequest{
				Name:      "shelves/1",
				PageSize:  100,
				PageToken: "token",
			},
			request2: &library.ListBooksRequest{
				Name:      "shelves/1",
				PageSize:  100,
				PageToken: "token",
			},
			equal: true,
		},
		{
			name: "different parents",
			request1: &library.ListBooksRequest{
				Name:      "shelves/1",
				PageSize:  100,
				PageToken: "token",
			},
			request2: &library.ListBooksRequest{
				Name:      "shelves/2",
				PageSize:  100,
				PageToken: "token",
			},
			equal: false,
		},
		{
			name: "different page sizes",
			request1: &library.ListBooksRequest{
				Name:      "shelves/1",
				PageSize:  100,
				PageToken: "token",
			},
			request2: &library.ListBooksRequest{
				Name:      "shelves/1",
				PageSize:  200,
				PageToken: "token",
			},
			equal: true,
		},
		{
			name: "different page tokens",
			request1: &library.ListBooksRequest{
				Name:      "shelves/1",
				PageSize:  100,
				PageToken: "token",
			},
			request2: &library.ListBooksRequest{
				Name:      "shelves/1",
				PageSize:  100,
				PageToken: "token2",
			},
			equal: true,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checksum1, err := calculateRequestChecksum(tt.request1)
			assert.NilError(t, err)
			checksum2, err := calculateRequestChecksum(tt.request2)
			assert.NilError(t, err)
			if tt.equal {
				assert.Assert(t, checksum1 == checksum2)
			} else {
				assert.Assert(t, checksum1 != checksum2)
			}
		})
	}
}
