package aipreflect

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ResourceType represents a resource type name.
type ResourceType string // e.g. pubsub.googleapis.com/Topic.

// Validate checks that the resource type name is syntactically valid.
func (n ResourceType) Validate() error {
	if strings.Count(string(n), "/") != 1 {
		return fmt.Errorf("validate resource type name '%s': invalid format", n)
	}
	if err := validateServiceName(n.ServiceName()); err != nil {
		return fmt.Errorf("validate resource type name '%s': %w", n, err)
	}
	if err := validateType(n.Type()); err != nil {
		return fmt.Errorf("validate resource type name '%s': %w", n, err)
	}
	return nil
}

// ServiceName returns the service name of the resource type name.
func (n ResourceType) ServiceName() string {
	if i := strings.LastIndexByte(string(n), '/'); i >= 0 {
		return string(n[:i])
	}
	return ""
}

// Type returns the type of the resource type name.
func (n ResourceType) Type() string {
	if i := strings.LastIndexByte(string(n), '/'); i >= 0 {
		return string(n[i+1:])
	}
	return ""
}

func validateServiceName(serviceName string) error {
	if serviceName == "" {
		return fmt.Errorf("service name: empty")
	}
	if !strings.ContainsRune(serviceName, '.') {
		return fmt.Errorf("service name: must be a valid domain name")
	}
	return nil
}

func validateType(t string) error {
	if t == "" {
		return fmt.Errorf("type: is empty")
	}
	if firstRune, _ := utf8.DecodeRuneInString(t); !unicode.IsUpper(firstRune) {
		return fmt.Errorf("type: must start with an upper-case letter")
	}
	for _, r := range t {
		if !unicode.In(r, unicode.Letter, unicode.Digit) {
			return fmt.Errorf("type: must be UpperCamelCase")
		}
	}
	return nil
}
