package resourcename

// RangeParents iterates over all parents of the provided resource name.
// The iteration order is from root ancestor down to the closest parent.
// Collection segments are included in the iteration, to not require knowing the pattern.
// For full resource names, the service is omitted.
func RangeParents(name string, fn func(parent string) bool) {
	var sc Scanner
	sc.Init(name)
	// First segment: special-case to handle full resource names.
	if !sc.Scan() {
		return
	}
	start := sc.Start()
	if sc.End() != len(name) && !fn(name[start:sc.End()]) {
		return
	}
	// Scan remaining segments.
	for sc.Scan() {
		if sc.End() != len(name) && !fn(name[start:sc.End()]) {
			return
		}
	}
}
