package fieldbehavior

import (
	"testing"

	examplefreightv1 "go.einride.tech/aip/proto/gen/einride/example/freight/v1"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/genproto/googleapis/example/library/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestClearFields(t *testing.T) {
	t.Parallel()
	t.Run("clear fields with set field_behavior", func(t *testing.T) {
		t.Parallel()
		site := &examplefreightv1.Site{
			Name:        "site1",           // has no field_behaviors; should not be cleared.
			CreateTime:  timestamppb.Now(), // has OUTPUT_ONLY field_behavior; should be cleared.
			DisplayName: "site one",        // has REQUIRED field_behavior; should not be cleared.
		}
		ClearFields(site, annotations.FieldBehavior_OUTPUT_ONLY)
		assert.Equal(t, site.GetCreateTime(), (*timestamppb.Timestamp)(nil))
		assert.Equal(t, site.GetDisplayName(), "site one")
		assert.Equal(t, site.GetName(), "site1")
	})
}

func TestCopyFields(t *testing.T) {
	t.Parallel()
	t.Run("different types", func(t *testing.T) {
		t.Parallel()
		assert.Assert(t, cmp.Panics(func() {
			CopyFields(&library.Book{}, &library.Shelf{}, annotations.FieldBehavior_REQUIRED)
		}))
	})
}

func TestValidateRequiredFields(t *testing.T) {
	t.Parallel()
	assert.NilError(t, ValidateRequiredFields(&examplefreightv1.GetShipmentRequest{Name: "testbook"}))
	assert.Error(t, ValidateRequiredFields(&examplefreightv1.GetShipmentRequest{}), "missing required field: name")
}

func TestValidateRequiredFieldsWithMask(t *testing.T) {
	t.Parallel()
	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		assert.NilError(
			t,
			ValidateRequiredFieldsWithMask(
				&library.Book{Name: "testbook"},
				nil,
			),
		)
	})
	t.Run("ok - empty mask", func(t *testing.T) {
		t.Parallel()
		assert.NilError(
			t,
			ValidateRequiredFieldsWithMask(
				&library.Book{},
				nil,
			),
		)
	})
	t.Run("missing field", func(t *testing.T) {
		t.Parallel()
		assert.Error(
			t,
			ValidateRequiredFieldsWithMask(
				&examplefreightv1.GetShipmentRequest{},
				&fieldmaskpb.FieldMask{Paths: []string{"*"}},
			),
			"missing required field: name",
		)
	})
	t.Run("missing but not in mask", func(t *testing.T) {
		t.Parallel()
		assert.NilError(
			t,
			ValidateRequiredFieldsWithMask(
				&library.Book{},
				&fieldmaskpb.FieldMask{Paths: []string{"author"}},
			),
		)
	})
	t.Run("missing nested", func(t *testing.T) {
		t.Parallel()
		assert.Error(
			t,
			ValidateRequiredFieldsWithMask(
				&examplefreightv1.UpdateShipmentRequest{
					Shipment: &examplefreightv1.Shipment{},
				},
				&fieldmaskpb.FieldMask{Paths: []string{"shipment.origin_site"}},
			),
			"missing required field: shipment.origin_site",
		)
	})
	t.Run("missing nested not in mask", func(t *testing.T) {
		t.Parallel()
		assert.NilError(
			t,
			ValidateRequiredFieldsWithMask(
				&library.UpdateBookRequest{
					Book: &library.Book{},
				},
				&fieldmaskpb.FieldMask{Paths: []string{"book.author"}},
			),
		)
	})
	t.Run("support maps", func(t *testing.T) {
		t.Parallel()
		assert.NilError(
			t,
			ValidateRequiredFieldsWithMask(
				&examplefreightv1.Shipment{
					Annotations: map[string]string{
						"x": "y",
					},
				},
				&fieldmaskpb.FieldMask{Paths: []string{"annotations"}},
			),
		)
	})
}

func TestValidateImmutableFieldsWithMask(t *testing.T) {
	t.Parallel()
	t.Run("no error when immutable field not set", func(t *testing.T) {
		t.Parallel()
		req := &examplefreightv1.UpdateShipmentRequest{
			Shipment: &examplefreightv1.Shipment{},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{""},
			},
		}
		err := ValidateImmutableFieldsWithMask(req, req.GetUpdateMask())
		assert.NilError(t, err)
	})
	t.Run("no error when immutable field not part of fieldmask", func(t *testing.T) {
		t.Parallel()
		req := &examplefreightv1.UpdateShipmentRequest{
			Shipment: &examplefreightv1.Shipment{
				ExternalReferenceId: "external-reference-id",
				OriginSite:          "shippers/shipper1/sites/site1",
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"shipment.origin_site"},
			},
		}
		err := ValidateImmutableFieldsWithMask(req, req.GetUpdateMask())
		assert.NilError(t, err)
	})
	t.Run("errors when wildcard fieldmask used", func(t *testing.T) {
		t.Parallel()
		req := &examplefreightv1.UpdateShipmentRequest{
			Shipment: &examplefreightv1.Shipment{},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"*"},
			},
		}
		err := ValidateImmutableFieldsWithMask(req, req.GetUpdateMask())
		assert.ErrorContains(t, err, "field is immutable")
	})
	t.Run("errors when immutable field set in fieldmask", func(t *testing.T) {
		t.Parallel()
		req := &examplefreightv1.UpdateShipmentRequest{
			Shipment: &examplefreightv1.Shipment{},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"shipment.external_reference_id"},
			},
		}
		err := ValidateImmutableFieldsWithMask(req, req.GetUpdateMask())
		assert.ErrorContains(t, err, "field is immutable")
	})
	t.Run("errors when immutable field set in message", func(t *testing.T) {
		t.Parallel()
		req := &examplefreightv1.UpdateShipmentRequest{
			Shipment: &examplefreightv1.Shipment{
				ExternalReferenceId: "I am immutable!",
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{},
			},
		}
		err := ValidateImmutableFieldsWithMask(req, req.GetUpdateMask())
		assert.ErrorContains(t, err, "field is immutable")
	})
	t.Run("errors when immutable field set in nested field", func(t *testing.T) {
		t.Parallel()
		req := &examplefreightv1.UpdateShipmentRequest{
			Shipment: &examplefreightv1.Shipment{
				LineItems: []*examplefreightv1.LineItem{
					{
						ExternalReferenceId: "I am immutable",
					},
				},
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{},
			},
		}
		err := ValidateImmutableFieldsWithMask(req, req.GetUpdateMask())
		assert.ErrorContains(t, err, "field is immutable")
	})
}
