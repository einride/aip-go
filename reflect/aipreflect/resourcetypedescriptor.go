package aipreflect

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ResourceType represents a resource type name.
type ResourceType string

// ResourceTypeDescriptor describes a resource type.
type ResourceTypeDescriptor struct {
	// ServiceName is the the name defined in the resource's service configuration.
	//
	// This usually (but not necessarily) matches the hostname that users use to call the service.
	//
	// For example: pubsub.googleapis.com.
	ServiceName string

	// Type is the type component of the resource type name.
	//
	// The type must be singular and use PascalCase (UpperCamelCase).
	Type string
}

// ParseName parses a resource type name string.
func NewResourceTypeDescriptor(s string) (ResourceTypeDescriptor, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return ResourceTypeDescriptor{}, fmt.Errorf("invalid format")
	}
	name := ResourceTypeDescriptor{
		ServiceName: parts[0],
		Type:        parts[1],
	}
	if err := name.Validate(); err != nil {
		return ResourceTypeDescriptor{}, err
	}
	return name, nil
}

// ResourceType returns the descriptor's resource type.
//
// For example: pubsub.googleapis.com/Topic.
func (n ResourceTypeDescriptor) ResourceType() ResourceType {
	return ResourceType(n.ServiceName + "/" + n.Type)
}

// String returns the string representation of the service type name.
func (n ResourceTypeDescriptor) String() string {
	return string(n.ResourceType())
}

// Validate the resource type name.
func (n ResourceTypeDescriptor) Validate() error {
	if err := n.validateServiceName(); err != nil {
		return fmt.Errorf("validate resource type name: %w", err)
	}
	if err := n.validateType(); err != nil {
		return fmt.Errorf("validate resource type name: %w", err)
	}
	return nil
}

func (n ResourceTypeDescriptor) validateServiceName() error {
	if n.ServiceName == "" {
		return fmt.Errorf("service name: empty")
	}
	if !strings.ContainsRune(n.ServiceName, '.') {
		return fmt.Errorf("service name: must be a valid domain name")
	}
	return nil
}

func (n ResourceTypeDescriptor) validateType() error {
	if n.Type == "" {
		return fmt.Errorf("type: is empty")
	}
	if firstRune, _ := utf8.DecodeRuneInString(n.Type); !unicode.IsUpper(firstRune) {
		return fmt.Errorf("type: must start with an upper-case letter")
	}
	if !isCamelCase(n.Type) {
		return fmt.Errorf("type: must be UpperCamelCase")
	}
	return nil
}

func isCamelCase(s string) bool {
	for _, r := range s {
		if !unicode.In(r, unicode.Letter, unicode.Digit) {
			return false
		}
	}
	return true
}
