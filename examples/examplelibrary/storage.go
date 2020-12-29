package examplelibrary

import (
	"context"

	"google.golang.org/genproto/googleapis/example/library/v1"
)

type Storage struct {
	Shelves []*library.Shelf
}

type ListShelvesQuery struct {
	Offset   int64
	PageSize int32
}

type ListShelvesResult struct {
	Shelves     []*library.Shelf
	HasNextPage bool
}

func (s *Storage) ListShelves(_ context.Context, query *ListShelvesQuery) (*ListShelvesResult, error) {
	pageStart := int(query.Offset)
	pageEnd := pageStart + int(query.PageSize)
	switch {
	case pageStart >= len(s.Shelves):
		return &ListShelvesResult{
			Shelves:     nil,
			HasNextPage: false,
		}, nil
	case pageEnd > len(s.Shelves):
		return &ListShelvesResult{
			Shelves:     s.Shelves[pageStart:],
			HasNextPage: false,
		}, nil
	default:
		return &ListShelvesResult{
			Shelves:     s.Shelves[pageStart:pageEnd],
			HasNextPage: true,
		}, nil
	}
}
