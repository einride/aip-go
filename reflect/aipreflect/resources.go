package aipreflect

import (
	"go.einride.tech/aip/resourcename"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// RangeResourceDescriptorsInFile iterates over all resource descriptors in a file while fn returns true.
// The iteration order is undefined.
func RangeResourceDescriptorsInFile(
	file protoreflect.FileDescriptor,
	fn func(resource *annotations.ResourceDescriptor) bool,
) {
	for _, resource := range proto.GetExtension(
		file.Options(), annotations.E_ResourceDefinition,
	).([]*annotations.ResourceDescriptor) {
		if !fn(resource) {
			return
		}
	}
	for i := 0; i < file.Messages().Len(); i++ {
		resource := proto.GetExtension(
			file.Messages().Get(i).Options(), annotations.E_Resource,
		).(*annotations.ResourceDescriptor)
		if resource == nil {
			continue
		}
		if !fn(resource) {
			return
		}
	}
}

// RangeResourceDescriptorsInPackage iterates over all resource descriptors in a package while fn returns true.
// The provided registry is used for looking up files in the package.
// The iteration order is undefined.
func RangeResourceDescriptorsInPackage(
	registry *protoregistry.Files,
	packageName protoreflect.FullName,
	fn func(resource *annotations.ResourceDescriptor) bool,
) {
	registry.RangeFilesByPackage(packageName, func(file protoreflect.FileDescriptor) bool {
		for _, resource := range proto.GetExtension(
			file.Options(), annotations.E_ResourceDefinition,
		).([]*annotations.ResourceDescriptor) {
			if !fn(resource) {
				return false
			}
		}
		for i := 0; i < file.Messages().Len(); i++ {
			resource := proto.GetExtension(
				file.Messages().Get(i).Options(), annotations.E_Resource,
			).(*annotations.ResourceDescriptor)
			if resource == nil {
				continue
			}
			if !fn(resource) {
				return false
			}
		}
		return true
	})
}

// RangeParentResourcesInPackage iterates over all a resource's parent descriptors in a package while fn returns true.
// The provided registry is used for looking up files in the package.
// The iteration order is undefined.
func RangeParentResourcesInPackage(
	registry *protoregistry.Files,
	packageName protoreflect.FullName,
	pattern string,
	fn func(parent *annotations.ResourceDescriptor) bool,
) {
	resourcename.RangeParents(pattern, func(parent string) bool {
		if parent == pattern {
			return true
		}
		var stop bool
		RangeResourceDescriptorsInPackage(registry, packageName, func(resource *annotations.ResourceDescriptor) bool {
			for _, candidatePattern := range resource.GetPattern() {
				if candidatePattern == parent {
					if !fn(resource) {
						stop = true
						return false
					}
					break
				}
			}
			return true
		})
		return !stop
	})
}
