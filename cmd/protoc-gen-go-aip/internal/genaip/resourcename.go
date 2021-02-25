package genaip

import (
	"strconv"
	"strings"

	"github.com/stoewer/go-strcase"
	"go.einride.tech/aip/reflect/aipreflect"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
)

type resourceNameCodeGenerator struct {
	resource *aipreflect.ResourceDescriptor
}

func newResourceNameCodeGenerator(resource *aipreflect.ResourceDescriptor) *resourceNameCodeGenerator {
	return &resourceNameCodeGenerator{resource: resource}
}

func (r *resourceNameCodeGenerator) GenerateCode(g *protogen.GeneratedFile) error {
	if len(r.resource.Names) == 0 {
		return nil
	}
	if len(r.resource.Names) > 1 || r.resource.History == annotations.ResourceDescriptor_FUTURE_MULTI_PATTERN {
		if err := r.generateInterface(g); err != nil {
			return err
		}
		if err := r.generateParseMethod(g); err != nil {
			return err
		}
	}
	if r.resource.History == annotations.ResourceDescriptor_HISTORY_UNSPECIFIED ||
		r.resource.History == annotations.ResourceDescriptor_ORIGINALLY_SINGLE_PATTERN {
		if err := r.generatePatternStruct(g, r.resource.Names[0], r.SinglePatternStructName()); err != nil {
			return err
		}
	}
	if r.resource.History == annotations.ResourceDescriptor_FUTURE_MULTI_PATTERN || len(r.resource.Names) > 1 {
		for _, name := range r.resource.Names {
			if err := r.generatePatternStruct(g, name, r.MultiPatternStructName(name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *resourceNameCodeGenerator) generatePatternStruct(
	g *protogen.GeneratedFile,
	name *aipreflect.ResourceNameDescriptor,
	typeName string,
) error {
	g.P()
	g.P("type ", typeName, " struct {")
	for _, segment := range name.Pattern.Segments {
		if segment.Variable {
			g.P(strcase.UpperCamelCase(segment.Value), " string")
		}
	}
	g.P("}")
	if err := r.generateValidateMethod(g, name, typeName); err != nil {
		return err
	}
	if err := r.generateStringMethod(g, name, typeName); err != nil {
		return err
	}
	if err := r.generateMarshalStringMethod(g, typeName); err != nil {
		return err
	}
	return r.generateUnmarshalStringMethod(g, name, typeName)
}

func (r *resourceNameCodeGenerator) generateValidateMethod(
	g *protogen.GeneratedFile,
	name *aipreflect.ResourceNameDescriptor,
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
	g.P("func (n *", typeName, ") Validate() error {")
	for _, segment := range name.Pattern.Segments {
		if !segment.Variable {
			continue
		}
		g.P("if n.", strcase.UpperCamelCase(segment.Value), " == ", strconv.Quote(""), "{")
		g.P("return ", fmtErrorf, `("`, segment.Value, `: empty")`)
		g.P("}")
		g.P("if ", stringsIndexByte, "(n.", strcase.UpperCamelCase(segment.Value), ", '/') != - 1 {")
		g.P("return ", fmtErrorf, `("`, segment.Value, `: contains illegal character '/'")`)
		g.P("}")
	}
	g.P("return nil")
	g.P("}")
	return nil
}

func (r *resourceNameCodeGenerator) generateStringMethod(
	g *protogen.GeneratedFile,
	name *aipreflect.ResourceNameDescriptor,
	typeName string,
) error {
	resourcenameSprint := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "go.einride.tech/aip/resourcename",
		GoName:       "Sprint",
	})
	g.P()
	g.P("func (n *", typeName, ") String() string {")
	g.P("return ", resourcenameSprint, "(")
	g.P(strconv.Quote(name.Pattern.String()), ",")
	for _, segment := range name.Pattern.Segments {
		if segment.Variable {
			g.P("n.", strcase.UpperCamelCase(segment.Value), ",")
		}
	}
	g.P(")")
	g.P("}")
	return nil
}

func (r *resourceNameCodeGenerator) generateMarshalStringMethod(
	g *protogen.GeneratedFile,
	typeName string,
) error {
	g.P()
	g.P("func (n *", typeName, ") MarshalString() (string, error) {")
	g.P("if err := n.Validate(); err != nil {")
	g.P("return ", strconv.Quote(""), ", err")
	g.P("}")
	g.P("return n.String(), nil")
	g.P("}")
	return nil
}

func (r *resourceNameCodeGenerator) generateUnmarshalStringMethod(
	g *protogen.GeneratedFile,
	name *aipreflect.ResourceNameDescriptor,
	typeName string,
) error {
	resourcenameSscan := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "go.einride.tech/aip/resourcename",
		GoName:       "Sscan",
	})
	g.P()
	g.P("func (n *", typeName, ") UnmarshalString(name string) error {")
	g.P("return ", resourcenameSscan, "(")
	g.P("name,")
	g.P(strconv.Quote(name.Pattern.String()), ",")
	for _, segment := range name.Pattern.Segments {
		if segment.Variable {
			g.P("&n.", strcase.UpperCamelCase(segment.Value), ",")
		}
	}
	g.P(")")
	g.P("}")
	return nil
}

func (r *resourceNameCodeGenerator) generateInterface(g *protogen.GeneratedFile) error {
	fmtStringer := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "fmt",
		GoName:       "Stringer",
	})
	g.P()
	g.P("type ", r.InterfaceName(), " interface {")
	g.P(fmtStringer)
	g.P("MarshalString() (string, error)")
	g.P("}")
	return nil
}

func (r *resourceNameCodeGenerator) generateParseMethod(g *protogen.GeneratedFile) error {
	resourcenameMatch := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "go.einride.tech/aip/resourcename",
		GoName:       "Match",
	})
	fmtErrorf := g.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: "fmt",
		GoName:       "Errorf",
	})
	g.P()
	g.P("func Parse", r.InterfaceName(), "(name string) (", r.InterfaceName(), ", error) {")
	g.P("switch {")
	for _, name := range r.resource.Names {
		g.P("case ", resourcenameMatch, "(", strconv.Quote(name.Pattern.String()), ", name):")
		g.P("var result ", r.PatternStructName(name))
		g.P("return &result, result.UnmarshalString(name)")
	}
	g.P("default:")
	g.P("return nil, ", fmtErrorf, `("no matching pattern")`)
	g.P("}")
	g.P("}")
	return nil
}

func (r *resourceNameCodeGenerator) PatternStructName(name *aipreflect.ResourceNameDescriptor) string {
	if len(r.resource.Names) >= 1 &&
		(r.resource.History == annotations.ResourceDescriptor_HISTORY_UNSPECIFIED ||
			r.resource.History == annotations.ResourceDescriptor_ORIGINALLY_SINGLE_PATTERN) &&
		r.resource.Names[0].Pattern.String() == name.Pattern.String() {
		return r.SinglePatternStructName()
	}
	return r.MultiPatternStructName(name)
}

func (r *resourceNameCodeGenerator) SinglePatternStructName() string {
	return r.resource.Type.Type() + "ResourceName"
}

func (r *resourceNameCodeGenerator) MultiPatternStructName(name *aipreflect.ResourceNameDescriptor) string {
	var result strings.Builder
	for _, segment := range name.Pattern.Segments {
		if !segment.Variable && aipreflect.GrammaticalName(segment.Value) != r.resource.Plural {
			_, _ = result.WriteString(strcase.UpperCamelCase(segment.Value))
		}
	}
	_, _ = result.WriteString(r.SinglePatternStructName())
	return result.String()
}

func (r *resourceNameCodeGenerator) InterfaceName() string {
	if r.resource.History == annotations.ResourceDescriptor_FUTURE_MULTI_PATTERN {
		return r.resource.Type.Type() + "ResourceName"
	}
	return r.resource.Type.Type() + "MultiPatternResourceName"
}
