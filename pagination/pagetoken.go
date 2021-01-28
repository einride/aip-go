package pagination

import (
	"fmt"
)

// PageToken is a page token that uses an offset to delineate which page to fetch.
type PageToken struct {
	// Offset of the page.
	Offset int64
	// RequestChecksum is the checksum of the request that generated the page token.
	RequestChecksum uint32
}

// pageTokenChecksumMask is a random bitmask applied to offset-based page token checksums.
//
// Change the bitmask to force checksum failures when changing the page token implementation.
const pageTokenChecksumMask uint32 = 0x9acb0442

// ParsePageToken parses an offset-based page token from the provided Request.
//
// If the request does not have a page token, a page token with offset 0 will be returned.
func ParsePageToken(request Request) (_ PageToken, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parse offset page token: %w", err)
		}
	}()
	requestChecksum, err := calculateRequestChecksum(request)
	if err != nil {
		return PageToken{}, err
	}
	requestChecksum ^= pageTokenChecksumMask // apply checksum mask for PageToken
	if request.GetPageToken() == "" {
		return PageToken{
			Offset:          0,
			RequestChecksum: requestChecksum,
		}, nil
	}
	var pageToken PageToken
	if err := DecodePageTokenStruct(request.GetPageToken(), &pageToken); err != nil {
		return PageToken{}, err
	}
	if pageToken.RequestChecksum != requestChecksum {
		return PageToken{}, fmt.Errorf(
			"checksum mismatch (got 0x%x but expected 0x%x)", pageToken.RequestChecksum, requestChecksum,
		)
	}
	return pageToken, nil
}

// Next returns the next page token for the provided Request.
func (p PageToken) Next(request Request) PageToken {
	p.Offset += int64(request.GetPageSize())
	return p
}

// String returns a string representation of the page token.
func (p PageToken) String() string {
	return EncodePageTokenStruct(&p)
}
