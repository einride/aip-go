package fieldmask

import "google.golang.org/protobuf/types/known/fieldmaskpb"

// WildcardPath is the path used for update masks that should perform a full replacement.
const WildcardPath = "*"

// IsFullReplacement reports whether a field mask contains the special wildcard path,
// meaning full replacement (the equivalent of PUT).
func IsFullReplacement(fm *fieldmaskpb.FieldMask) bool {
	return len(fm.GetPaths()) == 1 && fm.GetPaths()[0] == WildcardPath
}
