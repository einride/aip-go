package lint

import (
	"google.golang.org/protobuf/compiler/protogen"
)

// Problem contains information about a problem identified by a lint rules.
type Problem struct {
	// Message provides a short description of the problem.
	// This should be no more than a single sentence.
	Message string

	// Suggestion provides a suggested fix, if applicable.
	Suggestion string

	// RuleID is the ID of the lint rule that identified the problem.
	RuleID string

	// Location is the source location of the problem.
	Location protogen.Location
}
