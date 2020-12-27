package resourcetype

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ParseName parses a resource type name string.
func ParseName(s string) (Name, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return Name{}, fmt.Errorf("parse resource type name: invalid format")
	}
	name := Name{
		ServiceName: parts[0],
		Type:        parts[1],
	}
	if err := name.Validate(); err != nil {
		return Name{}, fmt.Errorf("parse resource type name: %w", err)
	}
	return name, nil
}

// Name represents a resource type name.
type Name struct {
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

// String returns the string representation of the service type name.
//
// For example: pubsub.googleapis.com/Topic.
func (n Name) String() string {
	return n.ServiceName + "/" + n.Type
}

// Validate the resource type name.
func (n Name) Validate() error {
	if err := n.validateServiceName(); err != nil {
		return fmt.Errorf("validate resource type name: %w", err)
	}
	if err := n.validateType(); err != nil {
		return fmt.Errorf("validate resource type name: %w", err)
	}
	return nil
}

func (n Name) validateServiceName() error {
	if n.ServiceName == "" {
		return fmt.Errorf("service name: empty")
	}
	if !strings.ContainsRune(n.ServiceName, '.') {
		return fmt.Errorf("service name: must be a valid domain name")
	}
	return nil
}

func (n Name) validateType() error {
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
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}
