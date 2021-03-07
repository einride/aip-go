package rules

import (
	"go.einride.tech/aip/lint"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

type AIP4231ResourceReferenceAnnotation struct{}

var (
	_ lint.Rule      = &AIP4231ResourceReferenceAnnotation{}
	_ lint.FieldRule = &AIP4231ResourceReferenceAnnotation{}
)

func (r *AIP4231ResourceReferenceAnnotation) RuleID() string {
	return "go.einride.tech/aip::4231::resource-reference-annotation"
}

func (r *AIP4231ResourceReferenceAnnotation) LintField(field *protogen.Field) ([]*lint.Problem, error) {
	resourceReference, ok := proto.GetExtension(
		field.Desc.Options(), annotations.E_ResourceReference,
	).(*annotations.ResourceReference)
	if !ok || resourceReference == nil {
		return nil, nil
	}
	switch {
	case resourceReference.GetType() == "" && resourceReference.GetChildType() == "":
		return []*lint.Problem{
			{Message: "must specify either type or child_type on resource_reference annotations"},
		}, nil
	case resourceReference.GetType() != "" && resourceReference.GetChildType() != "":
		return []*lint.Problem{
			{Message: "must not specify both type and child_type on resource_reference annotations"},
		}, nil
	default:
		return nil, nil
	}
}
