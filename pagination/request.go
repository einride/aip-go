package pagination

import (
	"fmt"
	"hash/crc32"

	"google.golang.org/protobuf/proto"
)

// Request is an interface for paginated request messages.
//
// See: https://google.aip.dev/158 (Pagination).
type Request interface {
	proto.Message
	// GetPageToken returns the page token of the request.
	GetPageToken() string
	// GetPageSize returns the page size of the request.
	GetPageSize() int32
}

type skipRequest interface {
	proto.Message
	// GetSkip returns the skip of the request.
	// See: https://google.aip.dev/158#skipping-results
	GetSkip() int32
}

// calculateRequestChecksum calculates a checksum for all fields of the request that must be the same across calls.
func calculateRequestChecksum(request Request) (uint32, error) {
	// Clone the original request, clear fields that may vary across calls, then checksum the resulting message.
	clonedRequest := proto.Clone(request)
	r := clonedRequest.ProtoReflect()
	r.Clear(r.Descriptor().Fields().ByName("page_token"))
	r.Clear(r.Descriptor().Fields().ByName("page_size"))
	if _, ok := request.(skipRequest); ok {
		r.Clear(r.Descriptor().Fields().ByName("skip"))
	}
	data, err := proto.Marshal(clonedRequest)
	if err != nil {
		return 0, fmt.Errorf("calculate request checksum: %w", err)
	}
	return crc32.ChecksumIEEE(data), nil
}
