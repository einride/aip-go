package aipreflect

import (
	"fmt"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ResourceDescriptor describes a resource.
type ResourceDescriptor struct {
	// ParentFile is the path of the parent file that the resource descriptor is declared in.
	ParentFile string
	// Message is the full name of the message that the resource descriptor is declared in.
	Message protoreflect.FullName
	// Type is the resource's type name.
	Type ResourceTypeName
	// Names are the resource name descriptors for the resource.
	Names []*ResourceNameDescriptor
	// Singular is the singular name of the resource type.
	Singular GrammaticalName
	// Plural is the plural name of the resource type.
	Plural GrammaticalName
	// Methods are the resource's known methods.
	Methods map[MethodType]protoreflect.FullName
}

// NewResourceDescriptor creates a new ResourceDescriptor from the provided resource descriptor message.
func NewResourceDescriptor(descriptor *annotations.ResourceDescriptor) (*ResourceDescriptor, error) {
	resource := &ResourceDescriptor{
		Type:     ResourceTypeName(descriptor.GetType()),
		Singular: GrammaticalName(descriptor.GetSingular()),
		Plural:   GrammaticalName(descriptor.GetPlural()),
	}
	if err := resource.Type.Validate(); err != nil {
		return nil, err
	}
	if resource.Singular != "" {
		if err := resource.Singular.Validate(); err != nil {
			return nil, err
		}
	}
	if resource.Plural != "" {
		if err := resource.Plural.Validate(); err != nil {
			return nil, err
		}
	}
	resource.Names = make([]*ResourceNameDescriptor, 0, len(descriptor.GetPattern()))
	for _, pattern := range descriptor.GetPattern() {
		resourceName, err := NewResourceNameDescriptor(pattern)
		if err != nil {
			return nil, err
		}
		resourceName.Type = resource.Type
		resource.Names = append(resource.Names, resourceName)
	}
	return resource, nil
}

// InferMethodName infers the method name of type t for the resource r.
func (r *ResourceDescriptor) InferMethodName(t MethodType) (protoreflect.Name, error) {
	if t.IsPlural() {
		if r.Plural == "" {
			return "", fmt.Errorf("infer %s method name %s: plural not specified", r.Type, t)
		}
		return protoreflect.Name(t) + protoreflect.Name(r.Plural.UpperCamelCase()), nil
	}
	if r.Singular == "" {
		return "", fmt.Errorf("infer %s method name %s: singular not specified", r.Type, t)
	}
	return protoreflect.Name(t) + protoreflect.Name(r.Singular.UpperCamelCase()), nil
}
