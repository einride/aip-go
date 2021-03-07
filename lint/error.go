package lint

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Error struct {
	Plugin   *protogen.Plugin
	Problems []*Problem
}

var _ error = &Error{}

func (r *Error) Error() string {
	var result strings.Builder
	for _, problem := range r.Problems {
		_ = result.WriteByte('\n')
		_, _ = result.WriteString(r.formatLocation(problem.Location))
		_, _ = result.WriteString(":\n")
		_ = result.WriteByte('\t')
		_, _ = result.WriteString(problem.RuleID)
		_ = result.WriteByte('\n')
		_ = result.WriteByte('\t')
		_, _ = result.WriteString(problem.Message)
		_ = result.WriteByte('\n')
		if problem.Suggestion != "" {
			_ = result.WriteByte('\t')
			_, _ = result.WriteString("Suggestion: ")
			_, _ = result.WriteString(problem.Suggestion)
			_ = result.WriteByte('\n')
		}
	}
	return result.String()
}

func (r *Error) formatLocation(location protogen.Location) string {
	file, ok := r.Plugin.FilesByPath[location.SourceFile]
	if !ok {
		return "<unknown-file>"
	}
	sourceLocations := file.Desc.SourceLocations()
	for i := 0; i < sourceLocations.Len(); i++ {
		sourceLocation := sourceLocations.Get(i)
		if equalSourcePaths(sourceLocation.Path, location.Path) {
			return fmt.Sprintf("%s:%d:%d", file.Desc.Path(), sourceLocation.StartLine+1, sourceLocation.StartColumn+1)
		}
	}
	return file.Desc.Path()
}

func equalSourcePaths(a, b protoreflect.SourcePath) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
