package aiptest

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"go.einride.tech/aip/internal/codegen"
	"go.einride.tech/aip/internal/protoloader"
	"go.einride.tech/aip/reflect/aipreflect"
	"go.einride.tech/aip/reflect/aipregistry"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func GeneratePackage(pkg PackageConfig) error {
	registry, err := protoloader.LoadFilesFromGoPackage(pkg.Path)
	if err != nil {
		return fmt.Errorf("load proto package: %w", err)
	}
	resources, err := aipregistry.NewResources(registry)
	if err != nil {
		return fmt.Errorf("load resources: %w", err)
	}
	for _, service := range pkg.Services {
		if err := generateService(service, resources, registry); err != nil {
			return fmt.Errorf("generate service '%s': %w", service.Name, err)
		}
	}
	return nil
}

func generateService(
	service ServiceConfig,
	aipreg *aipregistry.Resources,
	protoreg *protoregistry.Files,
) error {
	desc, err := findServiceDescriptor(protoreg, protoreflect.Name(service.Name))
	if err != nil {
		return err
	}
	resources, err := findServiceResources(aipreg, protoreflect.Name(service.Name))
	if err != nil {
		return err
	}
	filename := filepath.Join(service.Out.Path, "aiptest_test.go")
	f := codegen.NewFile(codegen.FileConfig{
		Filename:    filename,
		Package:     service.Out.Name,
		GeneratedBy: "aip-test",
	})

	for _, resource := range resources {
		msg, err := protoreg.FindDescriptorByName(resource.Message)
		if err != nil {
			return err
		}
		(&resourceGenerator{
			resource: resource,
			service:  desc,
			message:  msg.(protoreflect.MessageDescriptor),
		}).Generate(f)
	}
	content, err := f.Content()
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, content, 0o600); err != nil {
		log.Panic(err)
	}
	log.Println("wrote:", filename)
	return nil
}

func findServiceResources(
	resources *aipregistry.Resources,
	service protoreflect.Name,
) ([]*aipreflect.ResourceDescriptor, error) {
	var found []*aipreflect.ResourceDescriptor
	resources.RangeResources(func(descriptor *aipreflect.ResourceDescriptor) bool {
		for _, method := range descriptor.Methods {
			if method.Parent().Name() == service {
				found = append(found, descriptor)
				return true
			}
		}
		return true
	})
	if len(found) == 0 {
		return nil, fmt.Errorf("no resources found for service '%s'", service)
	}
	return found, nil
}

func findServiceDescriptor(
	registry *protoregistry.Files,
	service protoreflect.Name,
) (protoreflect.ServiceDescriptor, error) {
	var found []protoreflect.ServiceDescriptor
	registry.RangeFiles(func(descriptor protoreflect.FileDescriptor) bool {
		services := descriptor.Services()
		for i := 0; i < services.Len(); i++ {
			if services.Get(i).Name() == service {
				found = append(found, services.Get(i))
			}
		}
		return true
	})
	if len(found) == 0 {
		return nil, fmt.Errorf("no service named '%s' found", service)
	}
	if len(found) > 1 {
		return nil, fmt.Errorf("multiple services name '%s' found", service)
	}
	return found[0], nil
}
