package genaip

import (
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
)

// resourceMessageCodeGenerator generates resource-related methods on the proto message type itself.
type resourceMessageCodeGenerator struct {
	resource *annotations.ResourceDescriptor
	message  *protogen.Message
	file     *protogen.File
}

// GenerateCode generates ResourcePattern, ParentPattern, ResourceTypeName, SetName,
// and optionally request extractor methods (if the corresponding request types exist).
func (r resourceMessageCodeGenerator) GenerateCode(g *protogen.GeneratedFile) error {
	patterns := r.resource.GetPattern()
	if len(patterns) == 0 {
		return nil
	}
	pattern := patterns[0]
	parentPattern := deriveParentPattern(pattern)
	typeName := extractTypeName(r.resource.GetType())
	r.generateResourcePatternMethod(g, pattern)
	r.generateParentPatternMethod(g, parentPattern)
	r.generateResourceTypeNameMethod(g, typeName)
	r.generateSetNameMethod(g)
	// Only generate extractor methods if the corresponding request types exist.
	// This allows the codegen to work with proto files that only define resources
	// without the full AIP service definition.
	resourceName := r.message.GoIdent.GoName
	if r.hasMessageWithName("Create" + resourceName + "Request") {
		r.generateExtractFromCreateRequestMethod(g)
	}
	if r.hasMessageWithName("Update" + resourceName + "Request") {
		r.generateExtractFromUpdateRequestMethod(g)
	}
	return nil
}

// hasMessageWithName checks if a message with the given name exists in the file.
func (r resourceMessageCodeGenerator) hasMessageWithName(name string) bool {
	for _, msg := range r.file.Messages {
		if msg.GoIdent.GoName == name {
			return true
		}
	}
	return false
}

// generateResourcePatternMethod generates the ResourcePattern method.
func (r resourceMessageCodeGenerator) generateResourcePatternMethod(g *protogen.GeneratedFile, pattern string) {
	g.P()
	g.P("// ResourcePattern returns the resource name pattern for ", r.message.GoIdent.GoName, ".")
	g.P("func (*", r.message.GoIdent.GoName, ") ResourcePattern() string {")
	g.P("\treturn ", `"`, pattern, `"`)
	g.P("}")
}

// generateParentPatternMethod generates the ParentPattern method.
func (r resourceMessageCodeGenerator) generateParentPatternMethod(g *protogen.GeneratedFile, parentPattern string) {
	g.P()
	g.P("// ParentPattern returns the parent resource pattern for ", r.message.GoIdent.GoName, ".")
	g.P("// Returns empty string for top-level resources.")
	g.P("func (*", r.message.GoIdent.GoName, ") ParentPattern() string {")
	g.P("\treturn ", `"`, parentPattern, `"`)
	g.P("}")
}

// generateResourceTypeNameMethod generates the ResourceTypeName method.
func (r resourceMessageCodeGenerator) generateResourceTypeNameMethod(g *protogen.GeneratedFile, typeName string) {
	g.P()
	g.P("// ResourceTypeName returns the short type name for ", r.message.GoIdent.GoName, ".")
	g.P("func (*", r.message.GoIdent.GoName, ") ResourceTypeName() string {")
	g.P("\treturn ", `"`, typeName, `"`)
	g.P("}")
}

// generateSetNameMethod generates the SetName method.
func (r resourceMessageCodeGenerator) generateSetNameMethod(g *protogen.GeneratedFile) {
	g.P()
	g.P("// SetName sets the name field on ", r.message.GoIdent.GoName, ".")
	g.P("func (x *", r.message.GoIdent.GoName, ") SetName(name string) {")
	g.P("\tx.Name = name")
	g.P("}")
}

// generateExtractFromCreateRequestMethod generates the ExtractFromCreateRequest method.
// This follows AIP-133 conventions where Create<Resource>Request has a <resource> field.
//
// This method is only generated when the corresponding Create<Resource>Request type
// exists in the same file, ensuring the generated code compiles successfully.
func (r resourceMessageCodeGenerator) generateExtractFromCreateRequestMethod(g *protogen.GeneratedFile) {
	resourceName := r.message.GoIdent.GoName
	requestType := "Create" + resourceName + "Request"
	getterName := "Get" + resourceName
	protoMessage := g.QualifiedGoIdent(protogen.GoIdent{
		GoName:       "Message",
		GoImportPath: "google.golang.org/protobuf/proto",
	})
	g.P()
	g.P("// ExtractFromCreateRequest extracts the ", resourceName, " from a Create", resourceName, "Request.")
	g.P("// This follows AIP-133 conventions.")
	g.P("func (*", resourceName, ") ExtractFromCreateRequest(req ", protoMessage, ") ", protoMessage, " {")
	g.P("\treturn req.(*", requestType, ").", getterName, "()")
	g.P("}")
}

// generateExtractFromUpdateRequestMethod generates the ExtractFromUpdateRequest method.
// This follows AIP-134 conventions where Update<Resource>Request has a <resource> field.
//
// This method is only generated when the corresponding Update<Resource>Request type
// exists in the same file, ensuring the generated code compiles successfully.
func (r resourceMessageCodeGenerator) generateExtractFromUpdateRequestMethod(g *protogen.GeneratedFile) {
	resourceName := r.message.GoIdent.GoName
	requestType := "Update" + resourceName + "Request"
	getterName := "Get" + resourceName
	protoMessage := g.QualifiedGoIdent(protogen.GoIdent{
		GoName:       "Message",
		GoImportPath: "google.golang.org/protobuf/proto",
	})
	g.P()
	g.P("// ExtractFromUpdateRequest extracts the ", resourceName, " from an Update", resourceName, "Request.")
	g.P("// This follows AIP-134 conventions.")
	g.P("func (*", resourceName, ") ExtractFromUpdateRequest(req ", protoMessage, ") ", protoMessage, " {")
	g.P("\treturn req.(*", requestType, ").", getterName, "()")
	g.P("}")
}

// deriveParentPattern derives the parent pattern by removing the last collection/resource segment.
// Example: "organizations/{organization}/companies/{company}" -> "organizations/{organization}".
func deriveParentPattern(pattern string) string {
	segments := strings.Split(pattern, "/")
	if len(segments) < 2 {
		return ""
	}
	// Remove last two segments (collection and resource variable).
	parentSegments := segments[:len(segments)-2]
	return strings.Join(parentSegments, "/")
}

// extractTypeName extracts the short type name from a full resource type.
// Example: "example.com/Company" -> "Company".
func extractTypeName(fullType string) string {
	parts := strings.Split(fullType, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
