package pagination

import (
	"fmt"
)

// OffsetPageToken is a page token that uses an offset to delineate which page to fetch.
type OffsetPageToken struct {
	// Offset of the page.
	Offset int64
	// RequestChecksum is the checksum of the request that generated the page token.
	RequestChecksum uint32
}

// offsetPageTokenChecksumMask is a random bitmask applied to offset-based page token checksums.
//
// Change the bitmask to force checksum failures when changing the page token implementation.
const offsetPageTokenChecksumMask uint32 = 0x9acb0442

// ParseOffsetPageToken parses an offset-based page token from the provided Request.
//
// If the request does not have a page token, a page token with offset 0 will be returned.
func ParseOffsetPageToken(request Request) (_ OffsetPageToken, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parse offset page token: %w", err)
		}
	}()
	requestChecksum, err := calculateRequestChecksum(request)
	if err != nil {
		return OffsetPageToken{}, err
	}
	requestChecksum ^= offsetPageTokenChecksumMask // apply checksum mask for OffsetPageToken
	if request.GetPageToken() == "" {
		return OffsetPageToken{
			Offset:          0,
			RequestChecksum: requestChecksum,
		}, nil
	}
	var pageToken OffsetPageToken
	if err := gobDecode(request.GetPageToken(), &pageToken); err != nil {
		return OffsetPageToken{}, err
	}
	if pageToken.RequestChecksum != requestChecksum {
		return OffsetPageToken{}, fmt.Errorf(
			"checksum mismatch (got 0x%x but expected 0x%x)", pageToken.RequestChecksum, requestChecksum,
		)
	}
	return pageToken, nil
}

// Next returns the next page token for the provided Request.
func (p OffsetPageToken) Next(request Request) OffsetPageToken {
	p.Offset += int64(request.GetPageSize())
	return p
}

// String returns a string representation of the page token.
func (p OffsetPageToken) String() string {
	return gobEncode(&p)
}
