package lint

import (
	"google.golang.org/protobuf/compiler/protogen"
)

// Rule is an interface for lint rules.
type Rule interface {
	// RuleID returns the unique ID of the lint rules.
	RuleID() string
}

// InitializerRule is an interface for rules that require initialization.
type InitializerRule interface {
	Initialize(*protogen.Plugin) error
}

// FileRule is an interface for lint rules that concern files.
type FileRule interface {
	// LintFile lints a file.
	LintFile(*protogen.File) ([]*Problem, error)
}

// FieldRule is an interface for linters that concern fields.
type FieldRule interface {
	// LintField lints a field.
	LintField(*protogen.Field) ([]*Problem, error)
}

// MessageRule is an interface for linters that concern messages.
type MessageRule interface {
	// LintMessage lints a message.
	LintMessage(*protogen.Message) ([]*Problem, error)
}

// ServiceRule is an interface for linters that concern services.
type ServiceRule interface {
	// LintService lints a service.
	LintService(*protogen.Service) ([]*Problem, error)
}

// MethodRule is an interface for linters that concern methods.
type MethodRule interface {
	// LintMethod lints a method.
	LintMethod(*protogen.Method) ([]*Problem, error)
}

// EnumRule is an interface for linters that concern enums.
type EnumRule interface {
	// LintEnum lints an enum.
	LintEnum(*protogen.Enum) ([]*Problem, error)
}

// EnumValueRule is an interface for linters that concern enum values.
type EnumValueRule interface {
	// LintEnum lints an enum value.
	LintEnumValue(*protogen.EnumValue) ([]*Problem, error)
}
