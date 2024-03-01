package validation

import (
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error represents a message validation error.
type Error struct {
	fieldViolations []*errdetails.BadRequest_FieldViolation
	grpcStatus      *status.Status
	str             string
}

// NewError creates a new validation error from the provided field violations.
func NewError(fieldViolations []*errdetails.BadRequest_FieldViolation) error {
	if len(fieldViolations) == 0 {
		panic("validation.NewError: must provide at least one field violation")
	}
	return &Error{
		fieldViolations: fieldViolations,
	}
}

// GRPCStatus converts the validation error to a gRPC status with code INVALID_ARGUMENT.
func (e *Error) GRPCStatus() *status.Status {
	if e.grpcStatus == nil {
		var fields strings.Builder
		for i, fieldViolation := range e.fieldViolations {
			_, _ = fields.WriteString(fieldViolation.GetField())
			if i < len(e.fieldViolations)-1 {
				_, _ = fields.WriteString(", ")
			}
		}
		withoutDetails := status.Newf(codes.InvalidArgument, "invalid fields: %s", fields.String())
		if withDetails, err := withoutDetails.WithDetails(&errdetails.BadRequest{
			FieldViolations: e.fieldViolations,
		}); err != nil {
			e.grpcStatus = withoutDetails
		} else {
			e.grpcStatus = withDetails
		}
	}
	return e.grpcStatus
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.str == "" {
		if len(e.fieldViolations) == 1 {
			e.str = fmt.Sprintf(
				"field violation on %s: %s",
				e.fieldViolations[0].GetField(),
				e.fieldViolations[0].GetDescription(),
			)
		} else {
			var result strings.Builder
			_, _ = result.WriteString("field violation on multiple fields:\n")
			for i, fieldViolation := range e.fieldViolations {
				_, _ = result.WriteString(fmt.Sprintf(" | %s: %s", fieldViolation.GetField(), fieldViolation.GetDescription()))
				if i < len(e.fieldViolations)-1 {
					_ = result.WriteByte('\n')
				}
			}
			e.str = result.String()
		}
	}
	return e.str
}
