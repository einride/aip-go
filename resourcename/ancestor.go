package resourcename

// Ancestor extracts an ancestor from the provided name, using a pattern for the ancestor.
func Ancestor(name string, pattern string) (string, bool) {
	if name == "" || pattern == "" {
		return "", false
	}
	var scName, scPattern Scanner
	scName.Init(name)
	scPattern.Init(pattern)
	for scPattern.Scan() {
		if !scName.Scan() {
			return "", false
		}
		segment := scPattern.Segment()
		if segment.IsWildcard() {
			return "", false // wildcards not supported in patterns
		}
		if !segment.IsVariable() && segment != scName.Segment() {
			return "", false // not a match
		}
	}
	return name[:scName.end], true
}
