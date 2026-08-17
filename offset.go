package flop

import (
	"fmt"

	"github.com/ysomad/flop/filter"
)

// Params contains parameters for offset pagination.
type Params struct {
	Page   int32
	Offset int64
	Limit  int32

	orderBy OrderBy
	filter  filter.Filter
}

// OrderBy returns the resolved ordering associated with these parameters.
func (p Params) OrderBy() OrderBy { return p.orderBy.clone() }

// Filter returns the validated filter associated with these parameters.
func (p Params) Filter() filter.Filter { return p.filter }

// NewParams parses a one-based page number and page size into storage
// parameters. pageSize must already be normalized by the caller.
//
// A zero page number selects the first page.
func NewParams(page, limit int32) (Params, error) {
	if page < 0 {
		return Params{}, fmt.Errorf("flop: page number must not be negative")
	}
	if limit <= 0 {
		return Params{}, fmt.Errorf("flop: page size must be positive")
	}
	if page == 0 {
		page = 1
	}
	offset := int64(page-1) * int64(limit)
	return Params{
		Page:   page,
		Offset: offset,
		Limit:  limit,
	}, nil
}

// Page is a page returned by offset pagination.
type Page[T any] struct {
	Items      []T
	Num        int32
	TotalPages int64
	TotalItems int64
}

// NewPage creates a page from rows and the total number of matching
// items.
func NewPage[T any](rows []T, p Params, totalItems int64) (Page[T], error) {
	if p.Limit <= 0 || p.Page <= 0 || p.Offset < 0 || totalItems < 0 {
		return Page[T]{}, fmt.Errorf("flop: invalid offset parameters")
	}
	if int64(len(rows)) > int64(p.Limit) {
		return Page[T]{}, fmt.Errorf("flop: storage returned too many rows")
	}
	totalPages := totalItems / int64(p.Limit)
	if totalItems%int64(p.Limit) != 0 {
		totalPages++
	}
	return Page[T]{
		Items:      rows,
		Num:        p.Page,
		TotalPages: totalPages,
		TotalItems: totalItems,
	}, nil
}
