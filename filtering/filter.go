package filtering

import (
	"fmt"

	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"
)

// Filter represents a parsed and type-checked filter.
type Filter struct {
	CheckedExpr *expr.CheckedExpr
}

// ApplyMacros modifies the filter by applying the provided macros.
func (f *Filter) ApplyMacros(macros ...Macro) error {
	typeMap := applyMacros(f.CheckedExpr.GetExpr(), f.CheckedExpr.GetSourceInfo(), macros...)
	for id, newType := range typeMap {
		oldType, ok := f.CheckedExpr.TypeMap[id]
		if ok && !proto.Equal(oldType, newType) {
			return fmt.Errorf("type conflict when applying macros. Expr ID %d defined with 2 different types: %s and %s", id, oldType, newType)
		}
		f.CheckedExpr.TypeMap[id] = newType
	}
	return nil
}
