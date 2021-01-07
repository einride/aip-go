package aipreflect

import (
	"fmt"
	"strings"
)

// ResourceNamePatternDescriptor describes a resource name pattern.
//
// Example: `publishers/{publisher}/books/{book}`.
type ResourceNamePatternDescriptor struct {
	// Segments are the individual segments in the pattern.
	Segments []ResourceNameSegmentDescriptor
}

// NewResourceNamePatternDescriptor creates a resource name pattern descriptorn from a resource name pattern.
//
// Pattern syntax from the documentation:
//
//   The path pattern must follow the syntax, which aligns with HTTP binding syntax:
//
//     Template = Segment { "/" Segment } ;
//     Segment = LITERAL | Variable ;
//     Variable = "{" LITERAL "}" ;
func NewResourceNamePatternDescriptor(pattern string) (ResourceNamePatternDescriptor, error) {
	if len(pattern) == 0 {
		return ResourceNamePatternDescriptor{}, fmt.Errorf("pattern is empty")
	}
	segments := strings.Split(pattern, "/")
	result := ResourceNamePatternDescriptor{
		Segments: make([]ResourceNameSegmentDescriptor, 0, len(segments)),
	}
	for _, value := range segments {
		if len(value) == 0 {
			return ResourceNamePatternDescriptor{}, fmt.Errorf("invalid format")
		}
		segment := ResourceNameSegmentDescriptor{
			Variable: value[0] == '{' && value[len(value)-1] == '}',
			Value:    value,
		}
		if segment.Variable {
			segment.Value = segment.Value[1 : len(segment.Value)-1]
		}
		result.Segments = append(result.Segments, segment)
	}
	return result, nil
}

// Parent returns a descriptor for the resource name's closest parent.
func (p ResourceNamePatternDescriptor) Parent() (ResourceNamePatternDescriptor, bool) {
	switch {
	case len(p.Segments) <= 2:
		return ResourceNamePatternDescriptor{}, false
	case p.IsSingleton():
		return ResourceNamePatternDescriptor{Segments: p.Segments[:len(p.Segments)-1]}, true
	default:
		return ResourceNamePatternDescriptor{Segments: p.Segments[:len(p.Segments)-2]}, true
	}
}

// Ancestors returns descriptors for the resource name's ancestors.
func (p ResourceNamePatternDescriptor) Ancestors() []ResourceNamePatternDescriptor {
	if len(p.Segments) <= 2 {
		return nil
	}
	result := make([]ResourceNamePatternDescriptor, 0, len(p.Segments)/2)
	for parent, ok := p.Parent(); ok; parent, ok = parent.Parent() {
		result = append(result, parent)
	}
	return result
}

// String returns the string representation of the resource name pattern.
func (p ResourceNamePatternDescriptor) String() string {
	var s strings.Builder
	s.Grow(p.Len())
	for i, segment := range p.Segments {
		if segment.Variable {
			_, _ = s.WriteRune('{')
			_, _ = s.WriteString(segment.Value)
			_, _ = s.WriteRune('}')
		} else {
			_, _ = s.WriteString(segment.Value)
		}
		if i < len(p.Segments)-1 {
			_, _ = s.WriteRune('/')
		}
	}
	return s.String()
}

// Len returns the length of the resource name pattern.
func (p ResourceNamePatternDescriptor) Len() int {
	result := len(p.Segments) - 1 // for the slashes
	for _, segment := range p.Segments {
		result += len(segment.Value) // for the value
		if segment.Variable {
			result += 2 // for the variable braces
		}
	}
	return result
}

// IsSingleton returns true if the pattern is a singleton pattern.
//
// From: https://aip.dev/156
//
//  Singleton resources must not have a user-provided or system-generated ID; their
//  resource name includes the name of their parent followed by one static-segment.
func (p ResourceNamePatternDescriptor) IsSingleton() bool {
	return len(p.Segments) > 2 && !p.Segments[len(p.Segments)-1].Variable
}

// IsAncestorOf returns true if p is an ancestor of child.
//
// For example, the pattern `publishers/{publisher}`
// is an ancestor of the pattern `publishers/{publisher}/books/{book}`.
func (p ResourceNamePatternDescriptor) IsAncestorOf(child ResourceNamePatternDescriptor) bool {
	if len(p.Segments) >= len(child.Segments) {
		return false
	}
	return segmentsEqual(p.Segments, child.Segments[:len(p.Segments)])
}

// NonVariableLen returns the non-variable length of the pattern, i.e. the length not counting variable segments.
//
// For example, the non-variable length of the pattern `resources/{resource}` is is 10.
func (p ResourceNamePatternDescriptor) NonVariableLen() int {
	result := len(p.Segments) - 1 // slashes
	for _, s := range p.Segments {
		if !s.Variable {
			result += len(s.Value)
		}
	}
	return result
}

// VariableCount returns the number of variables in the pattern.
func (p ResourceNamePatternDescriptor) VariableCount() int {
	var result int
	for _, s := range p.Segments {
		if s.Variable {
			result++
		}
	}
	return result
}

// ValidateResourceName validates a resource name against the pattern p.
func (p ResourceNamePatternDescriptor) ValidateResourceName(name string) error {
	nameSegmentsCount := strings.Count(name, "/") + 1
	if len(p.Segments) != nameSegmentsCount {
		return fmt.Errorf(
			"validate resource name `%s` against pattern `%s`: expected %d segments but got %d",
			name,
			p,
			len(p.Segments),
			nameSegmentsCount,
		)
	}
	remainingName := name
	for i, segment := range p.Segments {
		indexOfNextSlash := strings.IndexRune(remainingName, '/')
		isFinalNameSegment := indexOfNextSlash == -1
		var currSegmentValue string
		if isFinalNameSegment {
			currSegmentValue = remainingName
			remainingName = ""
		} else {
			currSegmentValue = remainingName[:indexOfNextSlash]
			remainingName = remainingName[indexOfNextSlash+1:]
		}
		if segment.Variable {
			if currSegmentValue == "" {
				return fmt.Errorf(
					"validate resource name `%s` against pattern `%s`: segment {%s} is empty",
					name,
					p,
					segment.Value,
				)
			}
			continue
		}
		if segment.Value != currSegmentValue {
			return fmt.Errorf(
				"validate resource name `%s` against pattern `%s`: expected segment %d to be `%s` but got `%s`",
				name,
				p,
				i+1,
				segment.Value,
				currSegmentValue,
			)
		}
	}
	return nil
}

// MarshalResourceName marshals a resource name from the pattern p given a list of values for the variables.
func (p ResourceNamePatternDescriptor) MarshalResourceName(values ...string) (string, error) {
	variableCount := p.VariableCount()
	if len(values) != variableCount {
		return "", fmt.Errorf(
			"marshal resource name pattern `%s`: got %d values but expected %d",
			p,
			len(values),
			variableCount,
		)
	}
	var variableLen int
	for _, v := range values {
		variableLen += len(v)
	}
	var name strings.Builder
	name.Grow(p.NonVariableLen() + variableLen)
	var iValue int
	for iSegment, s := range p.Segments {
		if s.Variable {
			if values[iValue] == "" {
				return "", fmt.Errorf(
					"marshal resource name pattern `%s`: empty value for {%s}",
					p,
					s.Value,
				)
			}
			_, _ = name.WriteString(values[iValue])
			iValue++
		} else {
			name.WriteString(s.Value)
		}
		isLastSegment := iSegment == len(p.Segments)-1
		if !isLastSegment {
			_ = name.WriteByte('/')
		}
	}
	return name.String(), nil
}

// Wildcard returns a wildcard resource name representation of the pattern.
//
// For example, the wildcard representation of the pattern `resources/{resource}` is `resources/*`.
func (p ResourceNamePatternDescriptor) Wildcard() string {
	var parts []string
	for _, segment := range p.Segments {
		if segment.Variable {
			parts = append(parts, "*")
		} else {
			parts = append(parts, segment.Value)
		}
	}
	return strings.Join(parts, "/")
}

func segmentsEqual(s1, s2 []ResourceNameSegmentDescriptor) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
