package genaip

import (
	"strconv"
	"strings"

	"github.com/stoewer/go-strcase"
	"go.einride.tech/aip/reflect/aipreflect"
	"go.einride.tech/aip/resourcename"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type resourceNameCodeGenerator struct {
	resource *annotations.ResourceDescriptor
	file     *protogen.File
	files    *protoregistry.Files
}

func (r resourceNameCodeGenerator) GenerateCode(g *protogen.GeneratedFile) error {
	if len(r.resource.GetPattern()) == 0 {
		return nil
	}
	hasMultiPattern := len(r.resource.GetPattern()) > 1
	hasFutureMultiPattern := r.resource.GetHistory() == annotations.ResourceDescriptor_FUTURE_MULTI_PATTERN
	// Generate multi-pattern interface and parse methods if we have multiple patterns now or in the future.
	if hasMultiPattern || hasFutureMultiPattern {
		if err := r.generateMultiPatternInterface(g); err != nil {
			return err
		}
		if err := r.generateMultiPatternParseMethod(g); err != nil {
			return err
		}
	}

	// Generate the single-pattern struct unless we explicitly only want multi-patterns from the start.
	firstPattern := r.resource.GetPattern()[0]
	shouldGenerateSinglePatternStruct := !hasFutureMultiPattern
	firstSinglePatternStructName := r.SinglePatternStructName()
	if shouldGenerateSinglePatternStruct {
		if err := r.generatePatternStruct(
			g, firstPattern, firstSinglePatternStructName,
		); err != nil {
			return err
		}
	}
	// Generate the multi-pattern variant of the single-pattern struct if we need multi-pattern support.
	// If we've already generated single-pattern structs above, ignore top-level resources here as the
	// multi-pattern variant of the single-pattern struct will be identical to the single-pattern struct.
	firstMultiPatternStructName := r.MultiPatternStructName(firstPattern)
	equalMultiSinglePatternStructName := firstMultiPatternStructName == firstSinglePatternStructName
	avoidGeneratingSameSingleStruct := shouldGenerateSinglePatternStruct && equalMultiSinglePatternStructName
	if (hasMultiPattern || hasFutureMultiPattern) && !avoidGeneratingSameSingleStruct {
		if err := r.generatePatternStruct(
			g, firstPattern, firstMultiPatternStructName,
		); err != nil {
			return err
		}
	}
	// Generate multi-pattern structs for all but the first pattern.
	for _, pattern := range r.resource.GetPattern()[1:] {
		if err := r.generatePatternStruct(g, pattern, r.MultiPatternStructName(pattern)); err != nil {
			return err
		}
	}
	return nil
}

func (r resourceNameCodeGenerator) generatePatternStruct(
	g *protogen.GeneratedFile,
	pattern string,
	typeName string,
) error {
	g.P()
	g.P("type ", typeName, " struct {")
	var sc resourcename.Scanner
	sc.Init(pattern)
	for sc.Scan() {
		if sc.Segment().IsVariable() {
			g.P(strcase.UpperCamelCase(string(sc.Segment().Literal())), " string")
		}
	}
	g.P("}")
	var parentConstructorsErr error
	aipreflect.RangeParentResourcesInPackage(
		r.files,
		r.file.Desc.Package(),
		pattern,
		func(parent *annotations.ResourceDescriptor) bool {
			if err := r.generateParentConstructorMethod(g, pattern, typeName, parent); err != nil {
				parentConstructorsErr = err
				return false
			}
			return true
		},
	)
	if parentConstructorsErr != nil {
		return parentConstructorsErr
	}
	if err := r.generateValidateMethod(g, pattern, typeName); err != nil {
		return err
	}
	if err := r.generateContainsWildcardMethod(g, pattern, typeName); err != nil {
		return err
	}
	if err := r.generateStringMethod(g, pattern, typeName); err != nil {
		return err
	}
	if err := r.generateMarshalStringMethod(g, typeName); err != nil {
		return err
	}
	if err := r.generateUnmarshalStringMethod(g, pattern, typeName); err != nil {
		return err
	}
	var parentErr error
	aipreflect.RangeParentResourcesInPackage(
		r.files,
		r.file.Desc.Package(),
		pattern,
		func(parent *annotations.ResourceDescriptor) bool {
			if err := r.generateParentMethod(g, pattern, typeName, parent); err != nil {
				parentErr = err
				return false
			}
			return true
		},
	)
	return parentErr
}

func (r *resourceNameCodeGenerator) generateParentConstructorMethod(
	g *protogen.GeneratedFile,
	pattern string,
	typeName string,
	parent *annotations.ResourceDescriptor,
) error {
	var parentPattern string
	for _, parentCandidate := range parent.GetPattern() {
		if resourcename.HasParent(pattern, parentCandidate) {
			parentPattern = parentCandidate
			break
		}
	}
	if parentPattern == "" {
		return nil
	}
	pg := resourceNameCodeGenerator{resource: parent, file: r.file, files: r.files}
	parentStruct := pg.StructName(parentPattern)
	g.P()
	g.P("func (n ", parentStruct, ") ", typeName, "(")
	var sc resourcename.Scanner
	sc.Init(strings.TrimPrefix(pattern, parentPattern))
	for sc.Scan() {
		if sc.Segment().IsVariable() {
			g.P(strcase.LowerCamelCase(string(sc.Segment().Literal())), " string,")
		}
	}
	g.P(") ", typeName, " {")
	g.P("return ", typeName, "{")
	sc.Init(parentPattern)
	for sc.Scan() {
		if sc.Segment().IsVariable() {
			g.P(
				strcase.UpperCamelCase(string(sc.Segment().Literal())),
				": ",
				"n.",
				strcase.UpperCamelCase(string(sc.Segment().Literal())),
				",",
			)
		}
	}
	sc.Init(strings.TrimPrefix(pattern, parentPattern))
	for sc.Scan() {
		if sc.Segment().IsVariable() {
			g.P(
				strcase.UpperCamelCase(string(sc.Segment().Literal())),
				": ",
				strcase.LowerCamelCase(string(sc.Segment().Literal())),
				",",
			)
		}
	}
	g.P("}")
	g.P("}")
	return nil
}

func (r *resourceNameCodeGenerator) generateParentMethod(
	g *protogen.GeneratedFile,
	pattern string,
	typeName string,
	parent *annotations.ResourceDescriptor,
) error {
	var parentPattern string
	for _, parentCandidate := range parent.GetPattern() {
		if resourcename.HasParent(pattern, parentCandidate) {
			parentPattern = parentCandidate
			break
		}
	}
	if parentPattern == "" {
		return nil
	}
	pg := resourceNameCodeGenerator{resource: parent, file: r.file, files: r.files}
	parentStruct := pg.StructName(parentPattern)
	g.P()
	g.P("func (n ", typeName, ") ", parentStruct, "() ", parentStruct, " {")
	g.P("return ", parentStruct, "{")
	var sc resourcename.Scanner
	sc.Init(parentPattern)
	for sc.Scan() {
		if sc.Segment().IsVariable() {
			fieldName := strcase.UpperCamelCase(string(sc.Segment().Literal()))
			g.P(fieldName, ": n.", fieldName, ",")
		}
	}
	g.P("}")
	g.P("}")
	return nil
}

func (r resourceNameCodeGenerator) generateValidateMethod(
	g *protogen.GeneratedFile,
	pattern string,
	typeName string,
) error {
	stringsIndexByte := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "strings",
		GoName:       "IndexByte",
	})
	fmtErrorf := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "fmt",
		GoName:       "Errorf",
	})
	g.P()
	g.P("func (n ", typeName, ") Validate() error {")
	var sc resourcename.Scanner
	sc.Init(pattern)
	for sc.Scan() {
		if !sc.Segment().IsVariable() {
			continue
		}
		value := string(sc.Segment().Literal())
		g.P("if n.", strcase.UpperCamelCase(value), " == ", strconv.Quote(""), "{")
		g.P("return ", fmtErrorf, `("`, value, `: empty")`)
		g.P("}")
		g.P("if ", stringsIndexByte, "(n.", strcase.UpperCamelCase(value), ", '/') != - 1 {")
		g.P("return ", fmtErrorf, `("`, value, `: contains illegal character '/'")`)
		g.P("}")
	}
	g.P("return nil")
	g.P("}")
	return nil
}

func (r resourceNameCodeGenerator) generateContainsWildcardMethod(
	g *protogen.GeneratedFile,
	pattern string,
	typeName string,
) error {
	g.P()
	g.P("func (n ", typeName, ") ContainsWildcard() bool {")
	var returnStatement strings.Builder
	returnStatement.WriteString("return false")
	var sc resourcename.Scanner
	sc.Init(pattern)
	for sc.Scan() {
		if !sc.Segment().IsVariable() {
			continue
		}
		returnStatement.WriteString("|| n.")
		returnStatement.WriteString(strcase.UpperCamelCase(string(sc.Segment().Literal())))
		returnStatement.WriteString(" == ")
		returnStatement.WriteString(strconv.Quote("-"))
	}
	g.P(returnStatement.String())
	g.P("}")
	return nil
}

func (r *resourceNameCodeGenerator) generateStringMethod(
	g *protogen.GeneratedFile,
	pattern string,
	typeName string,
) error {
	resourcenameSprint := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "go.einride.tech/aip/resourcename",
		GoName:       "Sprint",
	})
	g.P()
	g.P("func (n ", typeName, ") String() string {")
	g.P("return ", resourcenameSprint, "(")
	g.P(strconv.Quote(pattern), ",")
	var sc resourcename.Scanner
	sc.Init(pattern)
	for sc.Scan() {
		if sc.Segment().IsVariable() {
			g.P("n.", strcase.UpperCamelCase(string(sc.Segment().Literal())), ",")
		}
	}
	g.P(")")
	g.P("}")
	return nil
}

func (r resourceNameCodeGenerator) generateMarshalStringMethod(
	g *protogen.GeneratedFile,
	typeName string,
) error {
	g.P()
	g.P("func (n ", typeName, ") MarshalString() (string, error) {")
	g.P("if err := n.Validate(); err != nil {")
	g.P("return ", strconv.Quote(""), ", err")
	g.P("}")
	g.P("return n.String(), nil")
	g.P("}")
	return nil
}

func (r resourceNameCodeGenerator) generateUnmarshalStringMethod(
	g *protogen.GeneratedFile,
	pattern string,
	typeName string,
) error {
	resourcenameSscan := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "go.einride.tech/aip/resourcename",
		GoName:       "Sscan",
	})
	g.P()
	g.P("func (n *", typeName, ") UnmarshalString(name string) error {")
	g.P("err := ", resourcenameSscan, "(")
	g.P("name,")
	g.P(strconv.Quote(pattern), ",")
	var sc resourcename.Scanner
	sc.Init(pattern)
	for sc.Scan() {
		if sc.Segment().IsVariable() {
			g.P("&n.", strcase.UpperCamelCase(string(sc.Segment().Literal())), ",")
		}
	}
	g.P(")")
	g.P("if err != nil {")
	g.P("return err")
	g.P("}")
	g.P("return n.Validate()")
	g.P("}")
	return nil
}

func (r resourceNameCodeGenerator) generateMultiPatternInterface(g *protogen.GeneratedFile) error {
	fmtStringer := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "fmt",
		GoName:       "Stringer",
	})
	g.P()
	g.P("type ", r.MultiPatternInterfaceName(), " interface {")
	g.P(fmtStringer)
	g.P("MarshalString() (string, error)")
	g.P("ContainsWildcard() bool")
	g.P("}")
	return nil
}

func (r *resourceNameCodeGenerator) generateMultiPatternParseMethod(g *protogen.GeneratedFile) error {
	resourcenameMatch := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "go.einride.tech/aip/resourcename",
		GoName:       "Match",
	})
	fmtErrorf := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "fmt",
		GoName:       "Errorf",
	})
	g.P()
	g.P("func Parse", r.MultiPatternInterfaceName(), "(name string) (", r.MultiPatternInterfaceName(), ", error) {")
	g.P("switch {")
	for _, pattern := range r.resource.GetPattern() {
		g.P("case ", resourcenameMatch, "(", strconv.Quote(pattern), ", name):")
		g.P("var result ", r.MultiPatternStructName(pattern))
		g.P("return &result, result.UnmarshalString(name)")
	}
	g.P("default:")
	g.P("return nil, ", fmtErrorf, `("no matching pattern")`)
	g.P("}")
	g.P("}")
	return nil
}

func (r *resourceNameCodeGenerator) SinglePatternStructName() string {
	return aipreflect.ResourceType(r.resource.GetType()).Type() + "ResourceName"
}

func (r *resourceNameCodeGenerator) StructName(pattern string) string {
	if r.resource.GetHistory() == annotations.ResourceDescriptor_FUTURE_MULTI_PATTERN || len(r.resource.GetPattern()) > 1 {
		return r.MultiPatternStructName(pattern)
	}
	if r.resource.GetPattern()[0] == pattern {
		return r.SinglePatternStructName()
	}
	return r.MultiPatternStructName(pattern)
}

func (r *resourceNameCodeGenerator) MultiPatternStructName(pattern string) string {
	var result strings.Builder
	var sc resourcename.Scanner
	sc.Init(pattern)
	for sc.Scan() {
		if !sc.Segment().IsVariable() && string(sc.Segment().Literal()) != r.resource.GetPlural() {
			_, _ = result.WriteString(strcase.UpperCamelCase(string(sc.Segment().Literal())))
		}
	}
	_, _ = result.WriteString(r.SinglePatternStructName())
	return result.String()
}

func (r *resourceNameCodeGenerator) MultiPatternInterfaceName() string {
	return aipreflect.ResourceType(r.resource.GetType()).Type() + "MultiPatternResourceName"
}
