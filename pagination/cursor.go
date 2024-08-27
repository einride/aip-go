package pagination

import "fmt"

// CursorToken is a page token that uses a cursor to delineate which page to fetch.
type CursorToken[C comparable] struct {
	// Cursor is the identifier of the fist element on the page.
	Cursor *C
	// https://google.aip.dev/158#skipping-results
	Skip int64
	// RequestChecksum is the checksum of the request that generated the page token.
	RequestChecksum uint32
}

// cursorChecksumMask is a random bitmask applied to cursor-based page token checksums.
//
// Change the bitmask to force checksum failures when changing the cursor implementation.
const cursorChecksumMask uint32 = 0x9acb0443

// ParseCursor parses a cursor-based page token from the provided Request.
//
// If the request does not have a page token, a page token with cursor set to nil will be returned.
func ParseCursor[C comparable](request Request) (CursorToken[C], error) {
	requestChecksum, err := calculateRequestChecksum(request)
	if err != nil {
		return CursorToken[C]{}, err
	}
	requestChecksum ^= cursorChecksumMask
	if request.GetPageToken() == "" {
		skip := int64(0)
		if s, ok := request.(skipRequest); ok {
			skip = int64(s.GetSkip())
		}
		return CursorToken[C]{
			Skip:            skip,
			RequestChecksum: requestChecksum,
		}, nil
	}
	var pageToken CursorToken[C]
	if err := DecodePageTokenStruct(request.GetPageToken(), &pageToken); err != nil {
		return CursorToken[C]{}, err
	}
	if pageToken.RequestChecksum != requestChecksum {
		return CursorToken[C]{}, fmt.Errorf(
			"checksum mismatch (got 0x%x but expected 0x%x)", pageToken.RequestChecksum, requestChecksum,
		)
	}
	if s, ok := request.(skipRequest); ok {
		pageToken.Skip = int64(s.GetSkip())
	}
	return pageToken, nil
}

func (c CursorToken[C]) String() string {
	return EncodePageTokenStruct(c)
}

// Next returns the next page token starting with the provided cursor.
func (c CursorToken[C]) Next(cursor C) CursorToken[C] {
	c.Cursor = &cursor
	c.Skip = 0
	return c
}
