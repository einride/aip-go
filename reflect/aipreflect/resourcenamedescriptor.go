package aipreflect

// ResourceNameDescriptor describes a resource name.
type ResourceNameDescriptor struct {
	// Type is the resource type name of the resource name's resource type.
	Type ResourceTypeName
	// Ancestors are the resource type names of the resource name's ancestors.
	Ancestors []ResourceTypeName
	// Pattern describes the resource name's pattern.
	Pattern ResourceNamePatternDescriptor
}

// NewResourceNameDescriptor returns a new resource name descriptor for the provided pattern.
func NewResourceNameDescriptor(pattern string) (*ResourceNameDescriptor, error) {
	patternDescriptor, err := NewResourceNamePatternDescriptor(pattern)
	if err != nil {
		return nil, err
	}
	return &ResourceNameDescriptor{
		Pattern: patternDescriptor,
	}, nil
}
