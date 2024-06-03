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
	t.Run("ok - optional primitive annotated with REQUIRED set to default value", func(t *testing.T) {
		t.Parallel()
		zero := int64(0)
		assert.NilError(
			t,
			ValidateRequiredFieldsWithMask(
				&examplefreightv1.Site{PersonnelCount: &zero},
				&fieldmaskpb.FieldMask{Paths: []string{"personnel_count"}},
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
	t.Run("missing optional field annotated with required", func(t *testing.T) {
		t.Parallel()
		assert.Error(
			t,
			ValidateRequiredFieldsWithMask(
				&examplefreightv1.Site{},
				&fieldmaskpb.FieldMask{Paths: []string{"personnel_count"}},
			),
			"missing required field: personnel_count",
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
