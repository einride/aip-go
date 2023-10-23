package resourcename

import (
	"fmt"
	"strings"
)

// Join combines resource names, separating them by slashes.
func Join(elems ...string) string {
	segments := make([]string, 0, len(elems))
	for elemIndex, elem := range elems {
		var sc Scanner
		sc.Init(elem)

		for sc.Scan() {
			if elemIndex == 0 && len(segments) == 0 && sc.Full() {
				segments = append(segments, fmt.Sprintf("//%s", sc.ServiceName()))
			}

			segment := sc.Segment()
			if segment == "" {
				continue
			}
			segments = append(segments, string(segment))
		}
	}

	if len(segments) == 0 {
		return "/"
	}

	return strings.Join(segments, "/")
}
