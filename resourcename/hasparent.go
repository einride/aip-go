package resourcename

// HasParent tests whether name has the specified parent. Wildcard segments (-) are considered.
func HasParent(name, parent string) bool {
	if name == "" || parent == "" {
		return false // empty is never a valid child or parent
	}
	var parentScanner, nameScanner Scanner
	parentScanner.Init(parent)
	nameScanner.Init(name)
	for parentScanner.Scan() {
		if !nameScanner.Scan() {
			return false
		}
		if parentScanner.Segment() != "-" && parentScanner.Segment() != nameScanner.Segment() {
			return false
		}
	}
	if parentScanner.Full() && nameScanner.Full() {
		return parentScanner.ServiceName() == nameScanner.ServiceName()
	}
	return true
}
