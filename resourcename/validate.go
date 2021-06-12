package resourcename

import (
	"fmt"
	"unicode"
)

// Validate that a resource name conforms to the restrictions outlined in AIP-122,
// primarily that each segment must be a valid DNS name.
// See: https://google.aip.dev/122
func Validate(name string) error {
	if name == "" {
		return fmt.Errorf("empty")
	}
	var sc Scanner
	sc.Init(name)
	var i int
	for sc.Scan() {
		i++
		segment := sc.Segment()
		switch {
		case segment == "":
			return fmt.Errorf("segment %d is empty", i)
		case segment == Wildcard:
			continue
		case segment.IsVariable():
			return fmt.Errorf("segment '%s': valid resource names must not contain variables", sc.Segment())
		case !isDomainName(string(sc.Segment())):
			return fmt.Errorf("segment '%s': not a valid DNS name", sc.Segment())
		}
	}
	if sc.Full() && !isDomainName(sc.ServiceName()) {
		return fmt.Errorf("service '%s': not a valid DNS name", sc.Segment())
	}
	return nil
}

// ValidatePattern that a resource name pattern conforms to the restrictions outlined in AIP-122,
// primarily that each segment must be a valid DNS name.
// See: https://google.aip.dev/122
func ValidatePattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("empty")
	}
	var sc Scanner
	sc.Init(pattern)
	var i int
	for sc.Scan() {
		i++
		segment := sc.Segment()
		switch {
		case segment == "":
			return fmt.Errorf("segment %d is empty", i)
		case segment == Wildcard:
			return fmt.Errorf("segment '%d': wildcards not allowed in patterns", i)
		case segment.IsVariable():
			switch {
			case segment.Literal() == "":
				return fmt.Errorf("segment '%s': missing variable name", sc.Segment())
			case !isSnakeCase(string(segment.Literal())):
				return fmt.Errorf("segment '%s': must be valid snake case", sc.Segment())
			}
		case !isDomainName(string(sc.Segment())):
			return fmt.Errorf("segment '%s': not a valid DNS name", sc.Segment())
		}
	}
	if sc.Full() {
		return fmt.Errorf("patterns can not be full resource names")
	}
	return nil
}

func isSnakeCase(s string) bool {
	for i, r := range s {
		if i == 0 {
			if !unicode.IsLower(r) {
				return false
			}
		} else if !(r == '_' || unicode.In(r, unicode.Lower, unicode.Digit)) {
			return false
		}
	}
	return true
}
