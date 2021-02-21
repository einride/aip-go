package aiptest

import (
	"strings"

	"go.einride.tech/aip/fieldbehavior"
	"go.einride.tech/aip/reflect/aipreflect"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func goImportPath(file protoreflect.FileDescriptor) []string {
	opts := file.Options().(*descriptorpb.FileOptions)
	opt := opts.GetGoPackage()
	if i := strings.IndexRune(opt, ';'); i != -1 {
		return []string{opt[:i], opt[i+1:]}
	}
	return []string{opt}
}

func serviceServer(service protoreflect.ServiceDescriptor) string {
	return string(service.Name()) + "Server"
}

func messageType(pkg string, msg protoreflect.MessageDescriptor) string {
	return pkg + "." + string(msg.Name())
}

func hasUserSettableID(resource *aipreflect.ResourceDescriptor, method protoreflect.MethodDescriptor) bool {
	idField := camelToSnake(resource.Singular.UpperCamelCase()) + "_id"
	return hasField(method.Input(), protoreflect.Name(idField))
}

func hasField(message protoreflect.MessageDescriptor, field protoreflect.Name) bool {
	f := message.Fields().ByName(field)
	return f != nil
}

func hasParent(resource *aipreflect.ResourceDescriptor) bool {
	if len(resource.Names) == 0 {
		return false
	}
	return len(resource.Names[0].Ancestors) > 0
}

func hasRequiredFields(message protoreflect.MessageDescriptor) bool {
	fields := message.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if fieldbehavior.Has(field, annotations.FieldBehavior_REQUIRED) {
			return true
		}
	}
	return false
}

func rangeRequiredFields(message protoreflect.MessageDescriptor, f func(field protoreflect.FieldDescriptor)) {
	fields := message.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if fieldbehavior.Has(field, annotations.FieldBehavior_REQUIRED) {
			f(field)
		}
	}
}
