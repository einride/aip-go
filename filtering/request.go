package filtering

// Request is an interface for gRPC requests that contain a standard AIP filter.
type Request interface {
	GetFilter() string
}

// ParseFilter parses and type-checks the filter in the provided Request.
func ParseFilter(request Request, declarations *Declarations) (Filter, error) {
	return ParseFilterString(request.GetFilter(), declarations)
}

// ParseFilter parses and type-checks the provided filter.
func ParseFilterString(filter string, declarations *Declarations) (Filter, error) {
	if filter == "" {
		return Filter{}, nil
	}
	var parser Parser
	parser.Init(filter)
	parsedExpr, err := parser.Parse()
	if err != nil {
		return Filter{}, err
	}
	var checker Checker
	checker.Init(parsedExpr.GetExpr(), parsedExpr.GetSourceInfo(), declarations)
	checkedExpr, err := checker.Check()
	if err != nil {
		return Filter{}, err
	}
	return Filter{
		CheckedExpr:  checkedExpr,
		declarations: declarations,
	}, nil
}
