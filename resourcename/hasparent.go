package resourcename

// HasParent tests whether name has the specified parent. Wildcard segments (-) are considered.
// Resource names without revisions are considered parents of the same resource name with a revision.
func HasParent(name, parent string) bool {
	if name == "" || parent == "" || name == parent {
		return false
	}
	var parentScanner, nameScanner Scanner
	parentScanner.Init(parent)
	nameScanner.Init(name)
	for parentScanner.Scan() {
		if !nameScanner.Scan() {
			return false
		}
		if parentScanner.Segment().IsWildcard() {
			continue
		}
		// Special-case: Identical resource IDs without revision are parents of revisioned resource IDs.
		if nameScanner.Segment().Literal().HasRevision() &&
			!parentScanner.Segment().Literal().HasRevision() &&
			(nameScanner.Segment().Literal().ResourceID() == parentScanner.Segment().Literal().ResourceID()) {
			continue
		}
		if parentScanner.Segment() != nameScanner.Segment() {
			return false
		}
	}
	if parentScanner.Full() && nameScanner.Full() {
		return parentScanner.ServiceName() == nameScanner.ServiceName()
	}
	return true
}
