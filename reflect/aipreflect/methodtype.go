package aipreflect

import "google.golang.org/protobuf/reflect/protoreflect"

// MethodType is an AIP method type.
type MethodType int

//go:generate stringer -type MethodType -trimprefix MethodType

const (
	// MethodTypeNone represents no method type.
	MethodTypeNone MethodType = iota

	// MethodTypeGet is the method type of the AIP standard Get method.
	// See: https://google.aip.dev/131 (Standard methods: Get).
	MethodTypeGet

	// MethodTypeList is the method type of the AIP standard List method.
	// See: https://google.aip.dev/132 (Standard methods: List).
	MethodTypeList

	// MethodTypeCreate is the method type of the AIP standard Create method.
	// See: https://google.aip.dev/133 (Standard methods: Create).
	MethodTypeCreate

	// MethodTypeUpdate is the method type of the AIP standard Update method.
	// See: https://google.aip.dev/133 (Standard methods: Update).
	MethodTypeUpdate

	// MethodTypeDelete is the method type of the AIP standard Delete method.
	// See: https://google.aip.dev/135 (Standard methods: Delete).
	MethodTypeDelete

	// MethodTypeUndelete is the method type of the AIP Undelete method for soft delete.
	// See: https://google.aip.dev/164 (Soft delete).
	MethodTypeUndelete

	// MethodTypeBatchGet is the method type of the AIP standard BatchGet method.
	// See: https://google.aip.dev/231 (Batch methods: Get).
	MethodTypeBatchGet

	// MethodTypeBatchCreate is the method type of the AIP standard BatchCreate method.
	// See: https://google.aip.dev/233 (Batch methods: Create).
	MethodTypeBatchCreate

	// MethodTypeBatchUpdate is the method type of the AIP standard BatchUpdate method.
	// See: https://google.aip.dev/234 (Batch methods: Update).
	MethodTypeBatchUpdate

	// MethodTypeBatchDelete is the method type of the AIP standard BatchDelete method.
	// See: https://google.aip.dev/235 (Batch methods: Delete).
	MethodTypeBatchDelete

	// MethodTypeSearch is the method type of the custom AIP method for searching a resource collection.
	// See: https://google.aip.dev/136 (Custom methods).
	MethodTypeSearch
)

// NamePrefix returns the method type's method name prefix.
func (s MethodType) NamePrefix() protoreflect.Name {
	return protoreflect.Name(s.String())
}

// IsPlural returns true if the method type relates to a plurality of resources.
func (s MethodType) IsPlural() bool {
	switch s {
	case MethodTypeList,
		MethodTypeSearch,
		MethodTypeBatchGet,
		MethodTypeBatchCreate,
		MethodTypeBatchUpdate,
		MethodTypeBatchDelete:
		return true
	case MethodTypeNone,
		MethodTypeGet,
		MethodTypeCreate,
		MethodTypeUpdate,
		MethodTypeDelete,
		MethodTypeUndelete:
		return false
	}
	return false
}
