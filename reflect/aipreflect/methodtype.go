package aipreflect

// MethodType is an AIP method type.
type MethodType string

const (
	// MethodTypeGet is the method type of the AIP standard Get method.
	// See: https://google.aip.dev/131 (Standard methods: Get).
	MethodTypeGet MethodType = "Get"

	// MethodTypeList is the method type of the AIP standard List method.
	// See: https://google.aip.dev/132 (Standard methods: List).
	MethodTypeList MethodType = "List"

	// MethodTypeCreate is the method type of the AIP standard Create method.
	// See: https://google.aip.dev/133 (Standard methods: Create).
	MethodTypeCreate MethodType = "Create"

	// MethodTypeUpdate is the method type of the AIP standard Update method.
	// See: https://google.aip.dev/133 (Standard methods: Update).
	MethodTypeUpdate MethodType = "Update"

	// MethodTypeDelete is the method type of the AIP standard Delete method.
	// See: https://google.aip.dev/135 (Standard methods: Delete).
	MethodTypeDelete MethodType = "Delete"

	// MethodTypeSearch is the method type of the custom AIP method for searching a resource collection.
	// See: https://google.aip.dev/136 (Custom methods).
	MethodTypeSearch MethodType = "Search"

	// MethodTypeUndelete is the method type of the AIP Undelete method for soft delete.
	// See: https://google.aip.dev/164 (Soft delete).
	MethodTypeUndelete MethodType = "Undelete"

	// MethodTypeBatchGet is the method type of the AIP standard BatchGet method.
	// See: https://google.aip.dev/231 (Batch methods: Get).
	MethodTypeBatchGet MethodType = "BatchGet"

	// MethodTypeBatchCreate is the method type of the AIP standard BatchCreate method.
	// See: https://google.aip.dev/233 (Batch methods: Create).
	MethodTypeBatchCreate MethodType = "BatchCreate"

	// MethodTypeBatchUpdate is the method type of the AIP standard BatchUpdate method.
	// See: https://google.aip.dev/234 (Batch methods: Update).
	MethodTypeBatchUpdate MethodType = "BatchUpdate"

	// MethodTypeBatchDelete is the method type of the AIP standard BatchDelete method.
	// See: https://google.aip.dev/235 (Batch methods: Delete).
	MethodTypeBatchDelete MethodType = "BatchDelete"
)

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
	case MethodTypeCreate,
		MethodTypeDelete,
		MethodTypeGet,
		MethodTypeUndelete,
		MethodTypeUpdate:
		return false
	}
	return false
}
