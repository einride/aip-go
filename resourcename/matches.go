package resourcename

// Matches reports whether the specified resource name matches the specified resource name pattern.
func Matches(name, pattern string) bool {
	var nameScanner, patternScanner Scanner
	nameScanner.Init(name)
	patternScanner.Init(pattern)
	for patternScanner.Scan() {
		if !nameScanner.Scan() {
			return false
		}
		nameSegment := nameScanner.Segment()
		if nameSegment.IsVariable() {
			return false
		}
		patternSegment := patternScanner.Segment()
		if patternSegment.IsWildcard() {
			return false // edge case - wildcard not allowed in patterns
		}
		if patternSegment.IsVariable() {
			if nameSegment == "" {
				return false
			}
		} else if nameSegment != patternSegment {
			return false
		}
	}
	switch {
	case
		nameScanner.Scan(),             // name has more segments than pattern, no match
		patternScanner.Segment() == "", // edge case - empty pattern never matches
		patternScanner.Full():          // edge case - full resource name not allowed in patterns
		return false
	}
	return true
}
