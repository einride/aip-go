package resourcename

import (
	"fmt"
	"strings"
)

// Pattern represents a resource name pattern.
//
// Example: `publishers/{publisher}/books/{book}`.
type Pattern struct {
	// StringVal is the literal string value of the pattern.
	StringVal string
	// Segments contains each Segment in the parsed pattern.
	Segments []Segment
}

// IsSingleton returns true if the pattern is a singleton pattern.
//
// From: https://aip.dev/156
//
//  Singleton resources must not have a user-provided or system-generated ID; their
//  resource name includes the name of their parent followed by one static-segment.
func (p Pattern) IsSingleton() bool {
	return len(p.Segments) > 2 && !p.Segments[len(p.Segments)-1].Variable
}

// IsAncestorOf returns true if p is an ancestor of child.
//
// For example, the pattern `publishers/{publisher}`
// is an ancestor of the pattern `publishers/{publisher}/books/{book}`.
func (p Pattern) IsAncestorOf(child Pattern) bool {
	if len(p.Segments) >= len(child.Segments) {
		return false
	}
	return segmentsEqual(p.Segments, child.Segments[:len(p.Segments)])
}

// NonVariableLen returns the non-variable length of the pattern, i.e. the length not counting variable segments.
//
// For example, the non-variable length of the pattern `resources/{resource}` is is 10.
func (p Pattern) NonVariableLen() int {
	result := len(p.Segments) - 1 // slashes
	for _, s := range p.Segments {
		if !s.Variable {
			result += len(s.Value)
		}
	}
	return result
}

// VariableCount returns the number of variables in the pattern.
func (p Pattern) VariableCount() int {
	var result int
	for _, s := range p.Segments {
		if s.Variable {
			result++
		}
	}
	return result
}

// MarshalResourceName marshals a resource name from the pattern p given a list of values for the variables.
func (p Pattern) MarshalResourceName(values ...string) (string, error) {
	variableCount := p.VariableCount()
	if len(values) != variableCount {
		return "", fmt.Errorf(
			"marshal resource name pattern `%s`: got %d values but expected %d",
			p.StringVal,
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
					p.StringVal,
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

// WildcardResourceName returns a wildcard resource name representation of the pattern.
//
// For example, the wildcard representation of the pattern `resources/{resource}` is `resources/*`.
func (p Pattern) WildcardResourceName() string {
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

// ParsePattern parses a resource name pattern string.
//
// Pattern syntax from the documentation:
//
//   The path pattern must follow the syntax, which aligns with HTTP binding syntax:
//
//     Template = Segment { "/" Segment } ;
//     Segment = LITERAL | Variable ;
//     Variable = "{" LITERAL "}" ;
func ParsePattern(s string) (Pattern, error) {
	if len(s) == 0 {
		return Pattern{}, fmt.Errorf("parse pattern: empty")
	}
	result := Pattern{StringVal: s}
	for _, value := range strings.Split(s, "/") {
		if len(value) == 0 {
			return Pattern{}, fmt.Errorf("parse pattern: invalid format")
		}
		segment := Segment{
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
