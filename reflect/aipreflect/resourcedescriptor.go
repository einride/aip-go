package aipreflect

import (
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
}

// NewResourceDescriptor creates a new ResourceDescriptor from the provided resource descriptor message.
func NewResourceDescriptor(descriptor *annotations.ResourceDescriptor) (*ResourceDescriptor, error) {
	resource := &ResourceDescriptor{
		Type: ResourceTypeName(descriptor.GetType()),
	}
	if err := resource.Type.Validate(); err != nil {
		return nil, err
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
