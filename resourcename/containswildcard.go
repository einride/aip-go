package resourcename

// ContainsWildcard reports whether the specified resource name contains any wildcard segments.
func ContainsWildcard(name string) bool {
	var sc Scanner
	sc.Init(name)
	for sc.Scan() {
		if sc.Segment().IsWildcard() {
			return true
		}
	}
	return false
}
