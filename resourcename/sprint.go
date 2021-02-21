package resourcename

import "strings"

// Sprintf formats resource name variables according to a pattern and returns the resulting string.
func Sprint(pattern string, variables ...string) string {
	var length, segments int
	var patternScanner Scanner
	patternScanner.Init(pattern)
	for patternScanner.Scan() {
		segment := patternScanner.Segment()
		if !segment.IsVariable() {
			length += len(segment.Literal())
		}
		segments++
	}
	for _, variable := range variables {
		length += len(variable)
	}
	if segments > 0 {
		length += segments - 1
	}
	var result strings.Builder
	result.Grow(length)
	patternScanner.Init(pattern)
	var i, variable int
	for patternScanner.Scan() {
		segment := patternScanner.Segment()
		if segment.IsVariable() {
			if variable < len(variables) {
				_, _ = result.WriteString(variables[variable])
				variable++
			}
		} else {
			_, _ = result.WriteString(string(segment.Literal()))
		}
		if i < segments-1 {
			_ = result.WriteByte('/')
		}
		i++
	}
	return result.String()
}
