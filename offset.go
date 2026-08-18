package flop

// OffsetParams contains parameters for offset pagination.
type OffsetParams struct {
	page   int32
	offset int64
	limit  int32

	orderBy OrderBy
	filter  Filter
}

// Page returns the one-based page number.
func (p OffsetParams) Page() int32 { return p.page }

// Offset returns the number of rows a storage implementation must skip.
func (p OffsetParams) Offset() int64 { return p.offset }

// Limit returns the page size.
func (p OffsetParams) Limit() int32 { return p.limit }

// OrderBy returns the resolved ordering associated with these parameters.
func (p OffsetParams) OrderBy() OrderBy { return p.orderBy.clone() }

// Filter returns the validated filter associated with these parameters.
func (p OffsetParams) Filter() Filter { return p.filter }

// NewOffsetParams parses a one-based page number and page size into storage
// parameters. orderBy and filter are carried through to storage unchanged.
// pageSize must already be normalized by the caller.
//
// A zero page number selects the first page.
func NewOffsetParams(page, pageSize int32, orderBy OrderBy, f Filter) (OffsetParams, error) {
	if page < 0 {
		return OffsetParams{}, errorf(ErrInvalidPage, "must not be negative")
	}
	if pageSize <= 0 {
		return OffsetParams{}, errorf(ErrInvalidPageSize, "must be positive")
	}
	if page == 0 {
		page = 1
	}
	offset := int64(page-1) * int64(pageSize)
	return OffsetParams{
		page:    page,
		offset:  offset,
		limit:   pageSize,
		orderBy: orderBy.clone(),
		filter:  f,
	}, nil
}

// OffsetPage is a page returned by offset pagination.
type OffsetPage[T any] struct {
	Items      []T
	Number     int32
	TotalPages int64
	TotalItems int64
}

// NewOffsetPage creates a page from rows and the total number of matching
// items.
func NewOffsetPage[T any](rows []T, p OffsetParams, totalItems int64) (OffsetPage[T], error) {
	if p.limit <= 0 || p.page <= 0 || p.offset < 0 || totalItems < 0 {
		return OffsetPage[T]{}, errorf(ErrDeclaration, "offset parameters are unset")
	}
	if int64(len(rows)) > int64(p.limit) {
		return OffsetPage[T]{}, errorf(ErrTooManyRows, "more rows than the page limit")
	}
	totalPages := totalItems / int64(p.limit)
	if totalItems%int64(p.limit) != 0 {
		totalPages++
	}
	return OffsetPage[T]{
		Items:      rows,
		Number:     p.page,
		TotalPages: totalPages,
		TotalItems: totalItems,
	}, nil
}
