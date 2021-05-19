package resourcename

import "strings"

// Parent returns the closest parent of the provided resource name or pattern.
// Returns empty string if there is no closest parent.
func Parent(name string) string {
	if strings.HasPrefix(name, "//") {
		firstIndexSlash := strings.IndexByte(name[2:], '/')
		if firstIndexSlash == -1 {
			return ""
		}
		name = name[2+firstIndexSlash+1:]
	}
	lastIndexSlash := strings.LastIndexByte(name, '/')
	if lastIndexSlash == -1 {
		return ""
	}
	return name[:lastIndexSlash]
}
