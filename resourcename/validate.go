package resourcename

import (
	"fmt"
)

// Validate that a resource name conforms to the restrictions outlined in AIP-122, primarily that each segment
// must be a valid DNS name.
// See: https://google.aip.dev/122
func Validate(name string) error {
	if name == "" {
		return fmt.Errorf("resource name is empty")
	}
	var sc Scanner
	sc.Init(name)
	var i int
	for sc.Scan() {
		i++
		switch {
		case sc.Segment() == "":
			return fmt.Errorf("segment %d is empty", i)
		case sc.Segment() == Wildcard:
			continue
		case !isDomainName(string(sc.Segment())):
			return fmt.Errorf("segment '%s': not a valid DNS name", sc.Segment())
		}
	}
	if sc.Full() && !isDomainName(sc.ServiceName()) {
		return fmt.Errorf("service '%s': not a valid DNS name", sc.Segment())
	}
	return nil
}
