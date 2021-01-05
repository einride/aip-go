package aipreflect

// ResourceNameDescriptor describes a resource name.
type ResourceNameDescriptor struct {
	Resource *ResourceDescriptor
	Parent   *ResourceDescriptor
	Pattern  *ResourceNamePatternDescriptor
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
