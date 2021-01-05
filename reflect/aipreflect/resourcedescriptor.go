package aipreflect

import (
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ResourceDescriptor describes a resource.
type ResourceDescriptor struct {
	// Descriptor is the original protobuf resource descriptor.
	Descriptor *annotations.ResourceDescriptor
	// File is the file that the resource descriptor is declared in.
	File protoreflect.FileDescriptor
	// Message is the message that the resource descriptor is declared in.
	Message protoreflect.MessageDescriptor
	// Type is the resource's type descriptor.
	Type ResourceTypeDescriptor
	// Names are the resource name descriptors for the resource.
	Names []*ResourceNameDescriptor
}

// NewResourceDescriptor creates a new ResourceDescriptor from the provided resource descriptor message.
func NewResourceDescriptor(descriptor *annotations.ResourceDescriptor) (*ResourceDescriptor, error) {
	resource := &ResourceDescriptor{
		Descriptor: descriptor,
	}
	resourceType, err := NewResourceTypeDescriptor(descriptor.GetType())
	if err != nil {
		return nil, err
	}
	resource.Type = resourceType
	resource.Names = make([]*ResourceNameDescriptor, 0, len(descriptor.GetPattern()))
	for _, pattern := range descriptor.GetPattern() {
		name, err := NewResourceNameDescriptor(pattern)
		if err != nil {
			return nil, err
		}
		name.Resource = resource
		resource.Names = append(resource.Names, name)
	}
	return resource, nil
}
