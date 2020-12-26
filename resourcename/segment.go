package resourcename

// Segment represents a segment in a resource name Pattern.
type Segment struct {
	Variable bool
	Value    string
}

func segmentsEqual(s1, s2 []Segment) bool {
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
