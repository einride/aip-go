package validation

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

// MessageValidator provides primitives for validating the fields of a message.
type MessageValidator struct {
	parentField     string
	fieldViolations []*errdetails.BadRequest_FieldViolation
}

// SetParentField sets a parent field which will be prepended to all the subsequently added violations.
func (m *MessageValidator) SetParentField(parentField string) {
	m.parentField = parentField
}

// AddFieldViolation adds a field violation to the message validator.
func (m *MessageValidator) AddFieldViolation(field, description string, formatArgs ...interface{}) {
	if m.parentField != "" {
		field = makeFieldWithParent(m.parentField, field)
	}
	if len(formatArgs) > 0 {
		description = fmt.Sprintf(description, formatArgs...)
	}
	m.fieldViolations = append(m.fieldViolations, &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: description,
	})
}

// AddFieldError adds a field violation from the provided error.
// If the provided error is a validation.Error, the individual field violations from the provided error are added.
func (m *MessageValidator) AddFieldError(field string, err error) {
	var errValidation *Error
	if errors.As(err, &errValidation) {
		// Add the child field violations with the current field as parent.
		originalParentField := m.parentField
		m.parentField = makeFieldWithParent(m.parentField, field)
		for _, fieldViolation := range errValidation.fieldViolations {
			m.AddFieldViolation(fieldViolation.GetField(), fieldViolation.GetDescription())
		}
		m.parentField = originalParentField
	} else {
		m.AddFieldViolation(field, err.Error())
	}
}

// Err returns the validator's current validation error, or nil if no field validations have been registered.
func (m *MessageValidator) Err() error {
	if len(m.fieldViolations) > 0 {
		return NewError(m.fieldViolations)
	}
	return nil
}

func makeFieldWithParent(parentField, field string) string {
	if parentField == "" {
		return field
	}
	var result strings.Builder
	result.Grow(len(parentField) + 1 + len(field))
	_, _ = result.WriteString(parentField)
	_ = result.WriteByte('.')
	_, _ = result.WriteString(field)
	return result.String()
}
