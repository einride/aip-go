package aipreflect

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestMethodType_IsPlural(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		methodType MethodType
		expected   bool
	}{
		{methodType: MethodTypeGet, expected: false},
		{methodType: MethodTypeCreate, expected: false},
		{methodType: MethodTypeDelete, expected: false},
		{methodType: MethodTypeGet, expected: false},
		{methodType: MethodTypeUndelete, expected: false},
		{methodType: MethodTypeUpdate, expected: false},
		{methodType: MethodTypeList, expected: true},
		{methodType: MethodTypeSearch, expected: true},
		{methodType: MethodTypeBatchGet, expected: true},
		{methodType: MethodTypeBatchCreate, expected: true},
		{methodType: MethodTypeBatchUpdate, expected: true},
		{methodType: MethodTypeBatchDelete, expected: true},
	} {
		assert.Assert(t, tt.methodType.IsPlural() == tt.expected, tt.methodType)
	}
}
