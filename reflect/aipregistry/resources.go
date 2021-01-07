package aipregistry

import (
	"fmt"

	"go.einride.tech/aip/reflect/aipreflect"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// Resources is a registry of resource descriptors and their relationships.
type Resources struct {
	resourcesByType         map[aipreflect.ResourceTypeName]*aipreflect.ResourceDescriptor
	resourcesByWildcardName map[string]*aipreflect.ResourceDescriptor
}

// NewResources creates a new resource registry from the provided file registry.
// The file registry is used to resolve parent / child relationships and to find standard methods.
func NewResources(files *protoregistry.Files) (*Resources, error) {
	registry := &Resources{
		resourcesByType:         make(map[aipreflect.ResourceTypeName]*aipreflect.ResourceDescriptor, files.NumFiles()),
		resourcesByWildcardName: make(map[string]*aipreflect.ResourceDescriptor, files.NumFiles()),
	}
	var responseErr error
	files.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		if err := registry.registerFile(file); err != nil {
			responseErr = err
			return false
		}
		return true
	})
	if responseErr != nil {
		return nil, responseErr
	}
	registry.inferAncestries()
	registry.inferMethods(files)
	return registry, nil
}

// FindDescriptorByName looks up a resource descriptor by the type name.
func (r *Resources) FindResourceByType(t aipreflect.ResourceTypeName) (*aipreflect.ResourceDescriptor, bool) {
	resource, ok := r.resourcesByType[t]
	return resource, ok
}

// RangeResources iterates over all registered resources while f returns true.
// The iteration order is undefined.
func (r *Resources) RangeResources(f func(*aipreflect.ResourceDescriptor) bool) {
	if r == nil {
		return
	}
	for _, resource := range r.resourcesByType {
		if !f(resource) {
			return
		}
	}
}

func (r *Resources) inferAncestries() {
	for _, resource := range r.resourcesByType {
		for _, name := range resource.Names {
			ancestorPatterns := name.Pattern.Ancestors()
			if len(ancestorPatterns) > 0 {
				name.Ancestors = make([]aipreflect.ResourceTypeName, 0, len(ancestorPatterns))
			}
			for _, ancestorPattern := range ancestorPatterns {
				ancestor, ok := r.resourcesByWildcardName[ancestorPattern.Wildcard()]
				if !ok {
					continue
				}
				ancestorResourceType := ancestor.Type
				name.Ancestors = append(name.Ancestors, ancestorResourceType)
			}
		}
	}
}

func (r *Resources) inferMethods(files *protoregistry.Files) {
	methodTypes := []aipreflect.MethodType{
		aipreflect.MethodTypeGet,
		aipreflect.MethodTypeList,
		aipreflect.MethodTypeCreate,
		aipreflect.MethodTypeUpdate,
		aipreflect.MethodTypeDelete,
		aipreflect.MethodTypeUndelete,
		aipreflect.MethodTypeBatchGet,
		aipreflect.MethodTypeBatchCreate,
		aipreflect.MethodTypeBatchUpdate,
		aipreflect.MethodTypeBatchDelete,
		aipreflect.MethodTypeSearch,
	}
	for _, resource := range r.resourcesByType {
		resource.Methods = make(map[aipreflect.MethodType]protoreflect.FullName)
		for _, methodType := range methodTypes {
			methodName, err := resource.InferMethodName(methodType)
			if err != nil {
				continue
			}
			file, err := files.FindFileByPath(resource.ParentFile)
			if err != nil {
				continue
			}
			resource, methodType := resource, methodType
			files.RangeFilesByPackage(file.Package(), func(packageFile protoreflect.FileDescriptor) bool {
				for i := 0; i < packageFile.Services().Len(); i++ {
					service := packageFile.Services().Get(i)
					for j := 0; j < service.Methods().Len(); j++ {
						method := service.Methods().Get(j)
						if method.Name() != methodName {
							continue
						}
						resource.Methods[methodType] = method.FullName()
						return false
					}
				}
				return true
			})
		}
	}
}

func (r *Resources) registerFile(file protoreflect.FileDescriptor) (err error) {
	descriptors := proto.GetExtension(file.Options(), annotations.E_ResourceDefinition)
	for _, descriptor := range descriptors.([]*annotations.ResourceDescriptor) {
		resource, err := aipreflect.NewResourceDescriptor(descriptor)
		if err != nil {
			return fmt.Errorf("register %s: %w", file.FullName(), err)
		}
		resource.ParentFile = file.Path()
		if err := r.registerResource(resource); err != nil {
			return fmt.Errorf("register %s: %w", file.FullName(), err)
		}
	}
	for i := 0; i < file.Messages().Len(); i++ {
		if err := r.registerMessage(file.Messages().Get(i)); err != nil {
			return fmt.Errorf("register %s: %w", file.FullName(), err)
		}
	}
	return nil
}

func (r *Resources) registerMessage(message protoreflect.MessageDescriptor) (err error) {
	descriptor := proto.GetExtension(message.Options(), annotations.E_Resource).(*annotations.ResourceDescriptor)
	if descriptor == nil {
		return nil
	}
	resource, err := aipreflect.NewResourceDescriptor(descriptor)
	if err != nil {
		return fmt.Errorf("register %s: %w", message.FullName(), err)
	}
	resource.ParentFile = message.ParentFile().Path()
	resource.Message = message.FullName()
	if err := r.registerResource(resource); err != nil {
		return fmt.Errorf("register %s: %w", message.FullName(), err)
	}
	return nil
}

func (r *Resources) registerResource(resource *aipreflect.ResourceDescriptor) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("register %s: %w", resource.Type, err)
		}
	}()
	if existingResource, ok := r.FindResourceByType(resource.Type); ok {
		switch {
		case resource.Message != "" && existingResource.Message != "":
			return fmt.Errorf(
				"conflict registering resource %s from message %s: already registered to message %s",
				resource.Type,
				resource.Message,
				existingResource.Message,
			)
		case resource.Message == "" && existingResource.Message != "":
			return nil // ignore file declarations when we already have the full message resource descriptor
		case resource.Message != "" && existingResource.Message == "":
			break // overwrite resource descriptors from file declarations with the full message descriptor
		case resource.Message == "" && existingResource.Message == "":
			if len(resource.Names) != len(existingResource.Names) {
				return fmt.Errorf(
					"conflict registering resource %s from file %s: file %s has registered with other resource names",
					resource.Type,
					resource.ParentFile,
					existingResource.ParentFile,
				)
			}
		}
	}
	r.resourcesByType[resource.Type] = resource
	for _, name := range resource.Names {
		r.resourcesByWildcardName[name.Pattern.Wildcard()] = resource
	}
	return nil
}
