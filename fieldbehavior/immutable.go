package fieldbehavior

// This file provides functions for validating immutable fields according to AIP-203.
//
// For AIP-203 compliant validation, use ValidateImmutableFieldsNotChanged, which allows
// immutable fields in the update mask as long as their values haven't changed.
//
// ValidateImmutableFieldsWithMask is deprecated as it's too strict and doesn't fully
// comply with AIP-203.

import (
	"fmt"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ValidateImmutableFieldsWithMask returns a validation error if the message
// or field mask contains a field that is immutable.
//
// Deprecated: This function is too strict and does not fully comply with AIP-203.
// It errors whenever an immutable field appears in the update mask, even if the value
// hasn't changed. According to AIP-203, immutable fields MAY be included in an update
// request as long as they are set to their existing value.
//
// Use ValidateImmutableFieldsNotChanged instead, which properly implements AIP-203 by
// only returning an error when an immutable field's value is actually being changed.
//
// If you want to ignore immutable fields rather than error then use ClearFields().
//
// See: https://aip.dev/203
func ValidateImmutableFieldsWithMask(m proto.Message, mask *fieldmaskpb.FieldMask) error {
	return validateImmutableFields(m.ProtoReflect(), mask, "")
}

func validateImmutableFields(m protoreflect.Message, mask *fieldmaskpb.FieldMask, path string) error {
	for i := 0; i < m.Descriptor().Fields().Len(); i++ {
		field := m.Descriptor().Fields().Get(i)
		currPath := path
		if len(currPath) > 0 {
			currPath += "."
		}

		currPath += string(field.Name())
		if isImmutable(field) && hasPath(mask, currPath) {
			return fmt.Errorf("field is immutable: %s", currPath)
		}

		if field.Kind() == protoreflect.MessageKind {
			value := m.Get(field)
			switch {
			case field.IsList():
				for i := 0; i < value.List().Len(); i++ {
					if err := validateImmutableFields(value.List().Get(i).Message(), mask, currPath); err != nil {
						return err
					}
				}
			case field.IsMap():
				if field.MapValue().Kind() != protoreflect.MessageKind {
					continue
				}
				var mapErr error
				value.Map().Range(func(_ protoreflect.MapKey, value protoreflect.Value) bool {
					if err := validateImmutableFields(value.Message(), mask, currPath); err != nil {
						mapErr = err
						return false
					}

					return true
				})
				if mapErr != nil {
					return mapErr
				}
			default:
				// Check if field is set to avoid infinite recursion on recursive type definitions
				if !m.Has(field) {
					continue
				}
				if err := validateImmutableFields(value.Message(), mask, currPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// ValidateImmutableFieldsNotChanged validates that immutable fields in the update mask
// are not being changed, in compliance with AIP-203.
//
// This function compares the old (existing) and updated (requested) values for any immutable
// fields present in the update mask. It returns a validation error only if an immutable
// field's value is actually being changed.
//
// According to AIP-203, immutable fields MAY be included in an update request as long as
// they are set to their existing value. This allows clients to use the same message for
// both create and update operations without having to remove immutable fields.
//
// Use this function when validating update requests to return INVALID_ARGUMENT if a user
// tries to change an immutable field's value.
//
// Example:
//
//	err := fieldbehavior.ValidateImmutableFieldsNotChanged(
//	    existingResource,
//	    updateRequest.GetResource(),
//	    updateRequest.GetUpdateMask(),
//	)
//	if err != nil {
//	    return status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
//	}
//
// See: https://aip.dev/203
func ValidateImmutableFieldsNotChanged(old, updated proto.Message, mask *fieldmaskpb.FieldMask) error {
	return validateImmutableFieldsNotChanged(old.ProtoReflect(), updated.ProtoReflect(), mask, "")
}

func validateImmutableFieldsNotChanged(
	old, updated protoreflect.Message,
	mask *fieldmaskpb.FieldMask,
	path string,
) error {
	if old.Descriptor() != updated.Descriptor() {
		return fmt.Errorf("old and updated messages have different types")
	}

	for i := 0; i < updated.Descriptor().Fields().Len(); i++ {
		field := updated.Descriptor().Fields().Get(i)
		currPath := path
		if len(currPath) > 0 {
			currPath += "."
		}

		currPath += string(field.Name())
		if isImmutable(field) && hasPathWithPrefix(mask, currPath) {
			// Check if the immutable field's value has changed
			oldValue := old.Get(field)
			updatedValue := updated.Get(field)
			if !oldValue.Equal(updatedValue) {
				return fmt.Errorf("immutable field cannot be changed: %s", currPath)
			}
			// Values are equal and entire field is in mask - no need to check nested fields
			continue
		}

		if field.Kind() == protoreflect.MessageKind {
			updatedValue := updated.Get(field)
			switch {
			case field.IsList():
				updatedList := updatedValue.List()
				oldList := old.Get(field).List()

				// If old is empty: new elements are allowed to have immutable fields set
				// If updated is empty: deletion is allowed
				if oldList.Len() == 0 || updatedList.Len() == 0 {
					continue
				}

				// Compare old and new lists element-by-element to catch nested immutable field changes
				// New elements are allowed to have immutable fields set, so only validate existing ones
				minLen := min(oldList.Len(), updatedList.Len())
				for i := range minLen {
					if err := validateImmutableFieldsNotChanged(
						oldList.Get(i).Message(),
						updatedList.Get(i).Message(),
						mask,
						currPath,
					); err != nil {
						return err
					}
				}

			case field.IsMap():
				if field.MapValue().Kind() != protoreflect.MessageKind {
					continue
				}

				updatedMap := updatedValue.Map()
				oldMap := old.Get(field).Map()

				// If old is empty: new map entries are allowed to have immutable fields set
				// If updated is empty: deletion is allowed
				if oldMap.Len() == 0 || updatedMap.Len() == 0 {
					continue
				}

				// Compare old and new maps key-by-key to catch nested immutable field changes
				var mapErr error
				updatedMap.Range(func(key protoreflect.MapKey, updatedVal protoreflect.Value) bool {
					if oldVal := oldMap.Get(key); oldVal.IsValid() {
						// Key exists in both maps: compare the values
						if err := validateImmutableFieldsNotChanged(
							oldVal.Message(),
							updatedVal.Message(),
							mask,
							currPath,
						); err != nil {
							mapErr = err
							return false
						}
					}
					// New keys are allowed to have immutable fields set, so skip validation
					return true
				})
				if mapErr != nil {
					return mapErr
				}

			default:
				// For singular message fields, check if set to avoid infinite recursion
				oldHas := old.Has(field)
				updatedHas := updated.Has(field)

				// If old is not set: new field is allowed to have immutable fields set
				// If updated is not set: deletion is allowed
				if !oldHas || !updatedHas {
					continue
				}

				// Both old and updated have this field set: compare nested messages
				if err := validateImmutableFieldsNotChanged(
					old.Get(field).Message(),
					updatedValue.Message(),
					mask,
					currPath,
				); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func isImmutable(field protoreflect.FieldDescriptor) bool {
	return Has(field, annotations.FieldBehavior_IMMUTABLE)
}
