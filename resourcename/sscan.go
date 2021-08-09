package resourcename

import (
	"fmt"
	"io"
)

// Sscan scans a resource name, storing successive segments into successive variables
// as determined by the provided pattern.
func Sscan(name, pattern string, variables ...*string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parse resource name '%s' with pattern '%s': %w", name, pattern, err)
		}
	}()
	var nameScanner, patternScanner Scanner
	nameScanner.Init(name)
	patternScanner.Init(pattern)
	var i int
	for patternScanner.Scan() {
		if patternScanner.Full() {
			return fmt.Errorf("invalid pattern")
		}
		if !nameScanner.Scan() {
			return fmt.Errorf("segment %s: %w", patternScanner.Segment(), io.ErrUnexpectedEOF)
		}
		nameSegment, patternSegment := nameScanner.Segment(), patternScanner.Segment()
		if !patternSegment.IsVariable() {
			if patternSegment.Literal() != nameSegment.Literal() {
				return fmt.Errorf("segment %s: got %s", patternSegment, nameSegment)
			}
			continue
		}
		if i > len(variables)-1 {
			return fmt.Errorf("segment %s: too few variables", patternSegment)
		}
		*variables[i] = string(nameSegment.Literal())
		i++
	}
	if nameScanner.Scan() {
		return fmt.Errorf("got trailing segments in name")
	}
	if i != len(variables) {
		return fmt.Errorf("too many variables: got %d but expected %d", i, len(variables))
	}
	return nil
}
