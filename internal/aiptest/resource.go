package aiptest

import (
	"strconv"

	"go.einride.tech/aip/internal/codegen"
	"go.einride.tech/aip/reflect/aipreflect"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type resourceGenerator struct {
	resource *aipreflect.ResourceDescriptor
	service  protoreflect.ServiceDescriptor
	message  protoreflect.MessageDescriptor
}

func (r *resourceGenerator) Generate(f *codegen.File) {
	r.generateFixture(f)
	cases := r.collectTestCases()
	r.generateRunner(f, cases)
	r.generateCtx(f)
	r.generateSkip(f)
	r.generateParentMethods(f)
	r.generateTestCases(f, cases)
}

func (r *resourceGenerator) generateFixture(f *codegen.File) {
	pkg := f.Import(goImportPath(r.service.ParentFile())...)
	context := f.Import("context")
	f.P("type ", r.fixtureName(), " struct {")

	// configuration fields
	f.P("Ctx ", context, ".Context")
	f.P("Service  ", pkg, ".", serviceServer(r.service))
	// TODO: don't always generate this
	_, hasCreate := r.resource.Methods[aipreflect.MethodTypeCreate]
	if hasParent(r.resource) {
		f.P("Parents []string")
		if hasCreate {
			f.P("Create func(parent string) *", messageType(pkg, r.message))
		}
	} else if hasCreate {
		f.P("Create func() *", messageType(pkg, r.message))
	}
	_, hasUpdate := r.resource.Methods[aipreflect.MethodTypeUpdate]
	if hasUpdate {
		if hasParent(r.resource) {
			f.P("Update func(parent string) *", messageType(pkg, r.message))
		} else {
			f.P("Update func() *", messageType(pkg, r.message))
		}
	}
	f.P("Skip []string")

	// fixture internal fields
	f.P("currParent int")
	f.P("}")
}

func (r *resourceGenerator) generateRunner(f *codegen.File, cases []testCase) {
	testingPkg := f.Import("testing")
	f.P("func (a *", r.fixtureName(), ") Test(t *", testingPkg, ".T) {")
	for _, tt := range cases {
		if !tt.enabled {
			continue
		}
		f.P("t.Run(", strconv.Quote(tt.Name()), ", a.", tt.FuncName(), ")")
	}
	f.P("}")
	f.P()
}

func (r *resourceGenerator) generateTestCases(f *codegen.File, cases []testCase) {
	testingPkg := f.Import("testing")
	for _, tt := range cases {
		if !tt.enabled {
			continue
		}
		f.P()
		f.P("func (a *", r.fixtureName(), ")", tt.FuncName(), "(t *", testingPkg, ".T) {")
		tt.fn(f)
		f.P("}")
	}
}

func (r *resourceGenerator) generateCtx(f *codegen.File) {
	ctx := f.Import("context")
	f.P("func (a *", r.fixtureName(), ") ctx() ", ctx, ".Context {")
	f.P("if a.Ctx == nil {")
	f.P("return ", ctx, ".Background()")
	f.P("}")
	f.P("return a.Ctx")
	f.P("}")
	f.P()
}

func (r *resourceGenerator) generateSkip(f *codegen.File) {
	testing := f.Import("testing")
	strings := f.Import("strings")
	f.P("func (a *", r.fixtureName(), ") maybeSkip(t *", testing, ".T) {")
	f.P("for _, skip := range a.Skip {")
	f.P("if ", strings, ".Contains(t.Name(), skip) {")
	f.P("t.Skip(\"skipped because of .Skip\")")
	f.P("}")
	f.P("}")
	f.P("}")
	f.P()
}

func (r *resourceGenerator) generateParentMethods(f *codegen.File) {
	if !hasParent(r.resource) {
		return
	}
	testing := f.Import("testing")
	f.P("func (a *", r.fixtureName(), ") nextParent(t *", testing, ".T, pristine bool) string {")
	f.P("if pristine {")
	f.P("a.currParent++")
	f.P("}")
	f.P("if a.currParent >= len(a.Parents) {")
	f.P("t.Fatal(\"need atleast\", a.currParent + 1,  \"parents\")")
	f.P("}")
	f.P("return a.Parents[a.currParent]")
	f.P("}")
	f.P()
	f.P("func (a *", r.fixtureName(), ") peekNextParent(t *", testing, ".T) string {")
	f.P("next := a.currParent + 1")
	f.P("if next >= len(a.Parents) {")
	f.P("t.Fatal(\"need atleast\", next +1,  \"parents\")")
	f.P("}")
	f.P("return a.Parents[next]")
	f.P("}")
	f.P()
}

func (r *resourceGenerator) collectTestCases() []testCase {
	return []testCase{
		r.createCase(),
		r.getCase(),
		r.batchGetCase(),
		r.updateCase(),
		r.listCase(),
		r.searchCase(),
	}
}

func (r *resourceGenerator) fixtureName() string {
	return "aipTest" + string(r.resource.Message.Name()) + "Fixture"
}
