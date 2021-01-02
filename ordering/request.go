package ordering

// Request is an interface for requests that support ordering.
//
// See: https://google.aip.dev/132#ordering (Standard methods: List > Ordering).
type Request interface {
	// GetOrderBy returns the ordering of the request.
	GetOrderBy() string
}

// ParseOrderBy request parses the ordering field for a Request.
func ParseOrderBy(r Request) (OrderBy, error) {
	var orderBy OrderBy
	if err := orderBy.UnmarshalString(r.GetOrderBy()); err != nil {
		return OrderBy{}, err
	}
	return orderBy, nil
}
