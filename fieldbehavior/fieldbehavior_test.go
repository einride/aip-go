package fieldbehavior

import (
	"testing"

	examplefreightv1 "go.einride.tech/aip/proto/gen/einride/example/freight/v1"
	syntaxv1 "go.einride.tech/aip/proto/gen/einride/example/syntax/v1"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/genproto/googleapis/example/library/v1"
	"google.golang.org/protobuf/testing/protocmp"
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
	t.Run("clear field with set field_behavior on nested message", func(t *testing.T) {
		t.Parallel()
		input := &syntaxv1.FieldBehaviorMessage{
			Field:         "field",    // has no field_behaviors; should not be cleared.
			OptionalField: "optional", // has OPTIONAL field_behavior; should not be cleared.
			MessageWithoutFieldBehavior: &syntaxv1.FieldBehaviorMessage{ // has no field_behaviors; should not be cleared.
				Field:           "field",       // has no field_behaviors; should not be cleared.
				OptionalField:   "optional",    // has OPTIONAL field_behavior; should not be cleared.
				OutputOnlyField: "output_only", // has OUTPUT_ONLY field_behavior; should be cleared.
			},
		}

		expected := &syntaxv1.FieldBehaviorMessage{
			Field:         "field",
			OptionalField: "optional",
			MessageWithoutFieldBehavior: &syntaxv1.FieldBehaviorMessage{
				Field:         "field",
				OptionalField: "optional",
			},
		}

		ClearFields(input, annotations.FieldBehavior_OUTPUT_ONLY)
		assert.DeepEqual(t, input, expected, protocmp.Transform())
	})

	t.Run("clear field with set field_behavior on multiple levels of nested messages", func(t *testing.T) {
		t.Parallel()
		input := &syntaxv1.FieldBehaviorMessage{
			Field:         "field",    // has no field_behaviors; should not be cleared.
			OptionalField: "optional", // has OPTIONAL field_behavior; should not be cleared.
			MessageWithoutFieldBehavior: &syntaxv1.FieldBehaviorMessage{ // has no field_behaviors; should not be cleared.
				Field:           "field",       // has no field_behaviors; should not be cleared.
				OptionalField:   "optional",    // has OPTIONAL field_behavior; should not be cleared.
				OutputOnlyField: "output_only", // has OUTPUT_ONLY field_behavior; should be cleared.
				MessageWithoutFieldBehavior: &syntaxv1.FieldBehaviorMessage{ // has no field_behaviors; should not be cleared.
					Field:           "field",       // has no field_behaviors; should not be cleared.
					OptionalField:   "optional",    // has OPTIONAL field_behavior; should not be cleared.
					OutputOnlyField: "output_only", // has OUTPUT_ONLY field_behavior; should be cleared.
				},
			},
		}

		expected := &syntaxv1.FieldBehaviorMessage{
			Field:         "field",
			OptionalField: "optional",
			MessageWithoutFieldBehavior: &syntaxv1.FieldBehaviorMessage{
				Field:         "field",
				OptionalField: "optional",
				MessageWithoutFieldBehavior: &syntaxv1.FieldBehaviorMessage{
					Field:         "field",
					OptionalField: "optional",
				},
			},
		}

		ClearFields(input, annotations.FieldBehavior_OUTPUT_ONLY)
		assert.DeepEqual(t, input, expected, protocmp.Transform())
	})

	t.Run("clear fields with set field_behavior on repeated message", func(t *testing.T) {
		t.Parallel()
		input := &syntaxv1.FieldBehaviorMessage{
			Field:         "field",    // has no field_behaviors; should not be cleared.
			OptionalField: "optional", // has OPTIONAL field_behavior; should not be cleared.
			RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{ // has no field_behaviors; should not be cleared.
				{
					Field:           "field",       // has no field_behaviors; should not be cleared.
					OptionalField:   "optional",    // has OPTIONAL field_behavior; should not be cleared.
					OutputOnlyField: "output_only", // has OUTPUT_ONLY field_behavior; should be cleared.
				},
			},
			StringList: []string{ // not a message type, should not be traversed
				"string",
			},
		}

		expected := &syntaxv1.FieldBehaviorMessage{
			Field:         "field",
			OptionalField: "optional",
			RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{
				{
					Field:         "field",
					OptionalField: "optional",
				},
			},
			StringList: []string{
				"string",
			},
		}

		ClearFields(input, annotations.FieldBehavior_OUTPUT_ONLY)
		assert.DeepEqual(t, input, expected, protocmp.Transform())
	})

	t.Run("clear fields with set field_behavior on multiple levels of repeated messages", func(t *testing.T) {
		t.Parallel()
		input := &syntaxv1.FieldBehaviorMessage{
			Field:         "field",    // has no field_behaviors; should not be cleared.
			OptionalField: "optional", // has OPTIONAL field_behavior; should not be cleared.
			RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{ // has no field_behaviors; should not be cleared.
				{
					Field:           "field",       // has no field_behaviors; should not be cleared.
					OptionalField:   "optional",    // has OPTIONAL field_behavior; should not be cleared.
					OutputOnlyField: "output_only", // has OUTPUT_ONLY field_behavior; should be cleared.
					RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{ // has no field_behaviors; should not be cleared.
						{
							Field:           "field",       // has no field_behaviors; should not be cleared.
							OptionalField:   "optional",    // has OPTIONAL field_behavior; should not be cleared.
							OutputOnlyField: "output_only", // has OUTPUT_ONLY field_behavior; should be cleared.
						},
					},
				},
			},
			StringList: []string{ // not a message type, should not be traversed
				"string",
			},
		}

		expected := &syntaxv1.FieldBehaviorMessage{
			Field:         "field",
			OptionalField: "optional",
			RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{
				{
					Field:         "field",
					OptionalField: "optional",
					RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{
						{
							Field:         "field",
							OptionalField: "optional",
						},
					},
				},
			},
			StringList: []string{
				"string",
			},
		}

		ClearFields(input, annotations.FieldBehavior_OUTPUT_ONLY)
		assert.DeepEqual(t, input, expected, protocmp.Transform())
	})

	t.Run("clear repeated field with set field_behavior", func(t *testing.T) {
		t.Parallel()
		input := &syntaxv1.FieldBehaviorMessage{
			Field:         "field",    // has no field_behaviors; should not be cleared.
			OptionalField: "optional", // has OPTIONAL field_behavior; should not be cleared.
			RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{ // has no field_behaviors; should not be cleared.
				{
					Field:         "field",    // has no field_behaviors; should not be cleared.
					OptionalField: "optional", // has OPTIONAL field_behavior; should not be cleared.
				},
			},
			RepeatedOutputOnlyMessage: []*syntaxv1.FieldBehaviorMessage{ // has OUTPUT_ONLY field_behavior; should be cleared.
				{
					Field:           "field",
					OptionalField:   "optional",
					OutputOnlyField: "output_only",
				},
			},
		}

		expected := &syntaxv1.FieldBehaviorMessage{
			Field:         "field",    // has no field_behaviors; should not be cleared.
			OptionalField: "optional", // has OPTIONAL field_behavior; should not be cleared.
			RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{ // has no field_behaviors; should not be cleared.
				{
					Field:         "field",    // has no field_behaviors; should not be cleared.
					OptionalField: "optional", // has OPTIONAL field_behavior; should not be cleared.
				},
			},
		}

		ClearFields(input, annotations.FieldBehavior_OUTPUT_ONLY)
		assert.DeepEqual(t, input, expected, protocmp.Transform())
	})

	t.Run("clear fields with set field_behavior on message in map", func(t *testing.T) {
		t.Parallel()
		input := &syntaxv1.FieldBehaviorMessage{
			OptionalField: "optional", // has OPTIONAL field_behavior; should not be cleared.
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{
				"key_1": {
					OptionalField:   "optional",    // has OPTIONAL field_behavior; should not be cleared.
					OutputOnlyField: "output_only", // has OUTPUT_ONLY field_behavior; should be cleared.
				},
			},
			StringMap: map[string]string{
				"string_key": "string", // not a message type, should not be traversed
			},
		}

		expected := &syntaxv1.FieldBehaviorMessage{
			OptionalField: "optional",
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{
				"key_1": {
					OptionalField: "optional",
				},
			},
			StringMap: map[string]string{
				"string_key": "string",
			},
		}

		ClearFields(input, annotations.FieldBehavior_OUTPUT_ONLY)
		assert.DeepEqual(
			t,
			input,
			expected,
			protocmp.Transform(),
		)
	})

	t.Run("clear map field with set field_behavior", func(t *testing.T) {
		t.Parallel()
		input := &syntaxv1.FieldBehaviorMessage{
			OptionalField: "optional",
			// has OUTPUT_ONLY field_behavior; should be cleared.
			MapOutputOnlyMessage: map[string]*syntaxv1.FieldBehaviorMessage{
				"key_1": {
					OutputOnlyField: "output_only", // has OUTPUT_ONLY field_behavior; but should be cleared with parent message.
				},
			},
		}

		expected := &syntaxv1.FieldBehaviorMessage{
			OptionalField: "optional",
		}

		ClearFields(input, annotations.FieldBehavior_OUTPUT_ONLY)
		assert.DeepEqual(
			t,
			input,
			expected,
			protocmp.Transform(),
		)
	})

	t.Run("clear field with set field_behavior on oneof message", func(t *testing.T) {
		t.Parallel()
		input := &syntaxv1.FieldBehaviorMessage{
			Field:         "field",    // has no field_behaviors; should not be cleared.
			OptionalField: "optional", // has OPTIONAL field_behavior; should not be cleared.
			Oneof: &syntaxv1.FieldBehaviorMessage_FieldBehaviorMessage{
				FieldBehaviorMessage: &syntaxv1.FieldBehaviorMessage{
					Field:           "field",       // has no field_behaviors; should not be cleared.
					OptionalField:   "optional",    // has OPTIONAL field_behavior; should not be cleared.
					OutputOnlyField: "output_only", // has OUTPUT_ONLY field_behavior; should be cleared.
				},
			},
		}

		expected := &syntaxv1.FieldBehaviorMessage{
			Field:         "field",
			OptionalField: "optional",
			Oneof: &syntaxv1.FieldBehaviorMessage_FieldBehaviorMessage{
				FieldBehaviorMessage: &syntaxv1.FieldBehaviorMessage{
					Field:         "field",
					OptionalField: "optional",
				},
			},
		}

		ClearFields(input, annotations.FieldBehavior_OUTPUT_ONLY)
		assert.DeepEqual(t, input, expected, protocmp.Transform())
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

func TestValidateImmutableFieldsNotChanged(t *testing.T) {
	t.Parallel()
	t.Run("no error when immutable field not in mask", func(t *testing.T) {
		t.Parallel()
		old := &examplefreightv1.Shipment{
			ExternalReferenceId: "external-reference-id",
			OriginSite:          "shippers/shipper1/sites/site1",
		}
		updated := &examplefreightv1.Shipment{
			ExternalReferenceId: "different-reference-id",
			OriginSite:          "shippers/shipper1/sites/site2",
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"origin_site"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("no error when immutable field unchanged", func(t *testing.T) {
		t.Parallel()
		old := &examplefreightv1.Shipment{
			ExternalReferenceId: "external-reference-id",
			OriginSite:          "shippers/shipper1/sites/site1",
		}
		updated := &examplefreightv1.Shipment{
			ExternalReferenceId: "external-reference-id",
			OriginSite:          "shippers/shipper1/sites/site2",
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"external_reference_id", "origin_site"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("error when immutable field changed", func(t *testing.T) {
		t.Parallel()
		old := &examplefreightv1.Shipment{
			ExternalReferenceId: "external-reference-id",
			OriginSite:          "shippers/shipper1/sites/site1",
		}
		updated := &examplefreightv1.Shipment{
			ExternalReferenceId: "different-reference-id",
			OriginSite:          "shippers/shipper1/sites/site2",
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"external_reference_id", "origin_site"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.ErrorContains(t, err, "immutable field cannot be changed")
		assert.ErrorContains(t, err, "external_reference_id")
	})
	t.Run("error when wildcard used and immutable field changed", func(t *testing.T) {
		t.Parallel()
		old := &examplefreightv1.Shipment{
			ExternalReferenceId: "external-reference-id",
			OriginSite:          "shippers/shipper1/sites/site1",
		}
		updated := &examplefreightv1.Shipment{
			ExternalReferenceId: "different-reference-id",
			OriginSite:          "shippers/shipper1/sites/site2",
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"*"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.ErrorContains(t, err, "immutable field cannot be changed")
	})
	t.Run("no error when wildcard used but immutable field unchanged", func(t *testing.T) {
		t.Parallel()
		old := &examplefreightv1.Shipment{
			ExternalReferenceId: "external-reference-id",
			OriginSite:          "shippers/shipper1/sites/site1",
		}
		updated := &examplefreightv1.Shipment{
			ExternalReferenceId: "external-reference-id",
			OriginSite:          "shippers/shipper1/sites/site2",
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"*"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("error when different message types", func(t *testing.T) {
		t.Parallel()
		old := &examplefreightv1.Shipment{
			ExternalReferenceId: "external-reference-id",
		}
		updated := &examplefreightv1.Site{
			Name: "site1",
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"*"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.ErrorContains(t, err, "different types")
	})
	t.Run("error when nested immutable field in repeated message changed", func(t *testing.T) {
		t.Parallel()
		old := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{
				{
					Title:               "Item 1",
					ExternalReferenceId: "line-item-1",
				},
			},
		}
		updated := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{
				{
					Title:               "Item 1 Updated",
					ExternalReferenceId: "line-item-1-changed",
				},
			},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"line_items"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.ErrorContains(t, err, "immutable field cannot be changed")
		assert.ErrorContains(t, err, "external_reference_id")
	})
	t.Run("no error when nested immutable field in repeated message unchanged", func(t *testing.T) {
		t.Parallel()
		old := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{
				{
					Title:               "Item 1",
					ExternalReferenceId: "line-item-1",
				},
			},
		}
		updated := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{
				{
					Title:               "Item 1 Updated",
					ExternalReferenceId: "line-item-1",
				},
			},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"line_items"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("error when new element added with immutable field", func(t *testing.T) {
		t.Parallel()
		old := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{
				{
					Title:               "Item 1",
					ExternalReferenceId: "line-item-1",
				},
			},
		}
		updated := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{
				{
					Title:               "Item 1",
					ExternalReferenceId: "line-item-1",
				},
				{
					Title:               "Item 2",
					ExternalReferenceId: "line-item-2",
				},
			},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"line_items"},
		}
		// This should succeed as we're adding new items, not changing existing ones
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("error when nested immutable field in map message changed", func(t *testing.T) {
		t.Parallel()
		old := &syntaxv1.FieldBehaviorMessage{
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{
				"key1": {
					Field:          "value1",
					ImmutableField: "immutable-1",
				},
			},
		}
		updated := &syntaxv1.FieldBehaviorMessage{
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{
				"key1": {
					Field:          "value1-updated",
					ImmutableField: "immutable-1-changed",
				},
			},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"map_optional_message"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.ErrorContains(t, err, "immutable field cannot be changed")
		assert.ErrorContains(t, err, "immutable_field")
	})
	t.Run("no error when nested immutable field in map message unchanged", func(t *testing.T) {
		t.Parallel()
		old := &syntaxv1.FieldBehaviorMessage{
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{
				"key1": {
					Field:          "value1",
					ImmutableField: "immutable-1",
				},
			},
		}
		updated := &syntaxv1.FieldBehaviorMessage{
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{
				"key1": {
					Field:          "value1-updated",
					ImmutableField: "immutable-1",
				},
			},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"map_optional_message"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("error when new map key added with changed immutable field", func(t *testing.T) {
		t.Parallel()
		old := &syntaxv1.FieldBehaviorMessage{
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{
				"key1": {
					Field:          "value1",
					ImmutableField: "immutable-1",
				},
			},
		}
		updated := &syntaxv1.FieldBehaviorMessage{
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{
				"key1": {
					Field:          "value1",
					ImmutableField: "immutable-1",
				},
				"key2": {
					Field:          "value2",
					ImmutableField: "immutable-2",
				},
			},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"map_optional_message"},
		}
		// This should succeed as we're adding new entries, not changing existing ones
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("no error when old list is empty and updated has elements with immutable fields", func(t *testing.T) {
		t.Parallel()
		old := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{},
		}
		updated := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{
				{
					Title:               "Item 1",
					ExternalReferenceId: "line-item-1",
				},
			},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"line_items"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("no error when updated list is empty and old had elements", func(t *testing.T) {
		t.Parallel()
		old := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{
				{
					Title:               "Item 1",
					ExternalReferenceId: "line-item-1",
				},
			},
		}
		updated := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"line_items"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("no error when both lists are empty", func(t *testing.T) {
		t.Parallel()
		old := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{},
		}
		updated := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"line_items"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("no error when old map is empty and updated has entries with immutable fields", func(t *testing.T) {
		t.Parallel()
		old := &syntaxv1.FieldBehaviorMessage{
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{},
		}
		updated := &syntaxv1.FieldBehaviorMessage{
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{
				"key1": {
					Field:          "value1",
					ImmutableField: "immutable-1",
				},
			},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"map_optional_message"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("no error when updated map is empty and old had entries", func(t *testing.T) {
		t.Parallel()
		old := &syntaxv1.FieldBehaviorMessage{
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{
				"key1": {
					Field:          "value1",
					ImmutableField: "immutable-1",
				},
			},
		}
		updated := &syntaxv1.FieldBehaviorMessage{
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"map_optional_message"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("no error when both maps are empty", func(t *testing.T) {
		t.Parallel()
		old := &syntaxv1.FieldBehaviorMessage{
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{},
		}
		updated := &syntaxv1.FieldBehaviorMessage{
			MapOptionalMessage: map[string]*syntaxv1.FieldBehaviorMessage{},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"map_optional_message"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("no error when old singular message field not set and updated has immutable field", func(t *testing.T) {
		t.Parallel()
		old := &syntaxv1.FieldBehaviorMessage{
			Field: "field",
		}
		updated := &syntaxv1.FieldBehaviorMessage{
			Field: "field",
			MessageWithoutFieldBehavior: &syntaxv1.FieldBehaviorMessage{
				ImmutableField: "immutable-value",
			},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"message_without_field_behavior"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("no error when updated singular message field not set and old had immutable field", func(t *testing.T) {
		t.Parallel()
		old := &syntaxv1.FieldBehaviorMessage{
			Field: "field",
			MessageWithoutFieldBehavior: &syntaxv1.FieldBehaviorMessage{
				ImmutableField: "immutable-value",
			},
		}
		updated := &syntaxv1.FieldBehaviorMessage{
			Field: "field",
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"message_without_field_behavior"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.NilError(t, err)
	})
	t.Run("error when trying to unset top-level immutable field", func(t *testing.T) {
		t.Parallel()
		old := &examplefreightv1.Shipment{
			ExternalReferenceId: "reference-123",
		}
		updated := &examplefreightv1.Shipment{
			ExternalReferenceId: "",
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"external_reference_id"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.ErrorContains(t, err, "immutable field cannot be changed")
	})
	t.Run("error when nested field under immutable parent is in mask and changed", func(t *testing.T) {
		t.Parallel()
		// The line_items field itself is NOT immutable, but line_items[].external_reference_id IS immutable
		// If mask contains "line_items.external_reference_id", we should still catch changes
		old := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{
				{
					Title:               "Item 1",
					ExternalReferenceId: "line-item-1",
				},
			},
		}
		updated := &examplefreightv1.Shipment{
			LineItems: []*examplefreightv1.LineItem{
				{
					Title:               "Item 1",
					ExternalReferenceId: "line-item-1-changed",
				},
			},
		}
		mask := &fieldmaskpb.FieldMask{
			Paths: []string{"line_items.external_reference_id"},
		}
		err := ValidateImmutableFieldsNotChanged(old, updated, mask)
		assert.ErrorContains(t, err, "immutable field cannot be changed")
		assert.ErrorContains(t, err, "external_reference_id")
	})
}
