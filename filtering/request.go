package filtering

import (
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Request is an interface for gRPC requests that contain a standard AIP filter.
type Request interface {
	GetFilter() string
}

// Filter represents a parsed and type-checked filter.
type Filter struct {
	CheckedExpr *expr.CheckedExpr
	EnumMap     map[int64]protoreflect.EnumDescriptor
}

// ParseFilter parses and type-checks the filter in the provided Request.
func ParseFilter(request Request, declarations *Declarations) (Filter, error) {
	if request.GetFilter() == "" {
		return Filter{}, nil
	}
	var parser Parser
	parser.Init(request.GetFilter())
	parsedExpr, err := parser.Parse()
	if err != nil {
		return Filter{}, err
	}
	var checker Checker
	checker.Init(parsedExpr, declarations)
	checkedExpr, err := checker.Check()
	if err != nil {
		return Filter{}, err
	}
	return Filter{
		CheckedExpr: checkedExpr,
	}, nil
}
