package aipmiddleware

import (
	"context"

	"go.einride.tech/aip/resourcename"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ParentValidatorConfig configures the ParentValidator middleware.
type ParentValidatorConfig struct {
	// ErrorLogHook is called when parent validation fails, with the failing request/response, parent/name combination.
	ErrorLogHook func(info *grpc.UnaryServerInfo, request, response proto.Message, parent, name string)
	// LogOnly configures the middleware to only log violations and not return errors.
	LogOnly bool
}

// ParentValidator validates that response resources are children of requested parent resources.
type ParentValidator struct {
	config ParentValidatorConfig
}

// NewParentValidator creates a new parent validator middleware.
func NewParentValidator(config ParentValidatorConfig) *ParentValidator {
	return &ParentValidator{config: config}
}

// UnaryServerInterceptor implements grpc.UnaryServerInterceptor.
func (p *ParentValidator) UnaryServerInterceptor(
	ctx context.Context,
	request interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	protoRequest, ok := request.(proto.Message)
	if !ok {
		return handler(ctx, request)
	}
	protoReflectRequest := protoRequest.ProtoReflect()
	// Request must have a parent field of string kind.
	parentField := protoReflectRequest.Descriptor().Fields().ByName("parent")
	if parentField == nil ||
		parentField.IsMap() ||
		parentField.IsList() ||
		parentField.Kind() != protoreflect.StringKind {
		return handler(ctx, request)
	}
	// The parent field must have a resource reference with a child type.
	parentResourceReference := proto.GetExtension(
		parentField.Options(), annotations.E_ResourceReference,
	).(*annotations.ResourceReference)
	if parentResourceReference == nil ||
		parentResourceReference.GetChildType() == "" && parentResourceReference.GetType() == "" {
		return handler(ctx, request)
	}
	parent := protoReflectRequest.Get(parentField).String()
	if parent == "" {
		return handler(ctx, request)
	}
	response, err := handler(ctx, request)
	if err != nil {
		return nil, err
	}
	protoResponse, ok := response.(proto.Message)
	if !ok {
		return response, err
	}
	if err := p.validateParent(
		info, protoRequest, protoResponse, parentResourceReference, parent,
	); err != nil && !p.config.LogOnly {
		return nil, err
	}
	return response, nil
}

func (p *ParentValidator) validateParent(
	info *grpc.UnaryServerInfo,
	request proto.Message,
	response proto.Message,
	parentResourceReference *annotations.ResourceReference,
	parent string,
) error {
	var result error
	response.ProtoReflect().Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		if field.Kind() != protoreflect.MessageKind || !field.IsList() {
			return true
		}
		fieldMessage := field.Message()
		fieldMessageNameField := fieldMessage.Fields().ByName("name")
		if fieldMessageNameField == nil || fieldMessageNameField.Kind() != protoreflect.StringKind {
			return true
		}
		resourceDescriptor := proto.GetExtension(
			fieldMessage.Options(), annotations.E_Resource,
		).(*annotations.ResourceDescriptor)
		if resourceDescriptor == nil {
			return true
		}
		if parentResourceReference.GetChildType() != "" &&
			resourceDescriptor.GetType() != parentResourceReference.GetChildType() {
			// Edge case: The response message contains a different type of resource than the specified child type.
			return true
		}
		listValue := value.List()
		for i := 0; i < listValue.Len(); i++ {
			name := listValue.Get(i).Message().Get(fieldMessageNameField).String()
			if !resourcename.HasParent(name, parent) {
				if p.config.ErrorLogHook != nil {
					p.config.ErrorLogHook(info, request, response, parent, name)
				}
				result = status.Error(codes.Internal, "parent validator failed")
				return false
			}
		}
		return true
	})
	return result
}
