package aiptest

import (
	"fmt"

	"go.einride.tech/aip/internal/codegen"
	"go.einride.tech/aip/reflect/aipreflect"
)

const defaultErr = "err"

type callCreateOpts struct {
	parent        string
	msg           string
	id            string
	response      string
	err           string
	assertNoError bool
	assign        string
}

func (r *resourceGenerator) callCreate(f *codegen.File, opts callCreateOpts) {
	pkg := f.Import(goImportPath(r.service.ParentFile())...)
	assert := f.Import("gotest.tools/v3/assert")
	createMethod, ok := r.resource.Methods[aipreflect.MethodTypeCreate]
	if !ok {
		// TODO: Support singleton resources (without a create method)
		panic(fmt.Errorf("no create method for resource '%s'", r.resource.Message))
	}
	if opts.err == "" {
		opts.err = defaultErr
	}
	if opts.assign == "" {
		opts.assign = ":="
	}
	create := r.service.Methods().ByName(createMethod.Name())
	f.P(opts.response, ", ", opts.err, " ", opts.assign, " a.Service.", createMethod.Name(), "(")
	f.P("a.ctx(),")
	f.P("&", pkg, ".", create.Input().Name(), "{")
	if hasParent(r.resource) {
		f.P("Parent: ", opts.parent, ",")
	}
	if opts.msg != "" {
		f.P(r.resource.Singular.UpperCamelCase(), ": ", opts.msg, ",")
	} else {
		if hasParent(r.resource) {
			f.P(r.resource.Singular.UpperCamelCase(), ": a.Create(", opts.parent, "),")
		} else {
			f.P(r.resource.Singular.UpperCamelCase(), ": a.Create(),")
		}
	}
	if hasUserSettableID(r.resource, create) && opts.id != "" {
		f.P(r.resource.Singular.UpperCamelCase(), "Id: ", opts.id, ",")
	}
	f.P("},")
	f.P(")")
	if opts.assertNoError {
		f.P(assert, ".NilError(t, ", opts.err, ")")
	}
}

type callGetOpts struct {
	response      string
	name          string
	err           string
	assertNoError bool
}

func (r *resourceGenerator) callGet(f *codegen.File, opts callGetOpts) {
	pkg := f.Import(goImportPath(r.service.ParentFile())...)
	assert := f.Import("gotest.tools/v3/assert")
	getMethod, ok := r.resource.Methods[aipreflect.MethodTypeGet]
	if !ok {
		panic(fmt.Errorf("no get method for resource '%s'", r.resource.Message))
	}
	if opts.err == "" {
		opts.err = defaultErr
	}
	get := r.service.Methods().ByName(getMethod.Name())
	f.P(opts.response, ", ", opts.err, " := a.Service.", getMethod.Name(), "(")
	f.P("a.ctx(),")
	f.P("&", pkg, ".", get.Input().Name(), "{")
	f.P("Name: ", opts.name, ",")
	f.P("},")
	f.P(")")
	if opts.assertNoError {
		f.P(assert, ".NilError(t, ", opts.err, ")")
	}
}

type callBatchGetOpts struct {
	parent        string
	names         []string
	response      string
	err           string
	assertNoError bool
}

func (r *resourceGenerator) callBatchGet(f *codegen.File, opts callBatchGetOpts) {
	pkg := f.Import(goImportPath(r.service.ParentFile())...)
	assert := f.Import("gotest.tools/v3/assert")
	batchGetMethod, ok := r.resource.Methods[aipreflect.MethodTypeBatchGet]
	if !ok {
		panic(fmt.Errorf("no batch get method for resource '%s'", r.resource.Message))
	}
	batchGet := r.service.Methods().ByName(batchGetMethod.Name())
	if opts.err == "" {
		opts.err = defaultErr
	}
	f.P(opts.response, ", err := a.Service.", batchGetMethod.Name(), "(")
	f.P("a.ctx(),")
	f.P("&", pkg, ".", batchGet.Input().Name(), "{")
	if hasParent(r.resource) {
		f.P("Parent: ", opts.parent, ",")
	}
	f.P("Names: []string{")
	for _, name := range opts.names {
		f.P(name, ",")
	}
	f.P("},")
	f.P("},")
	f.P(")")
	if opts.assertNoError {
		f.P(assert, ".NilError(t, ", opts.err, ")")
	}
}

type callUpdateOpts struct {
	parent          string
	name            string
	msg             string
	updateMaskPaths []string
	response        string
	err             string
	assertNoError   bool
	assign          string
}

func (r *resourceGenerator) callUpdate(f *codegen.File, opts callUpdateOpts) {
	pkg := f.Import(goImportPath(r.service.ParentFile())...)
	assert := f.Import("gotest.tools/v3/assert")
	fieldmask := f.Import("google.golang.org/protobuf/types/known/fieldmaskpb")
	updateMethod, ok := r.resource.Methods[aipreflect.MethodTypeUpdate]
	if !ok {
		panic(fmt.Errorf("no update method for resource '%s'", r.resource.Message))
	}
	if opts.err == "" {
		opts.err = defaultErr
	}
	if opts.assign == "" {
		opts.assign = ":="
	}
	update := r.service.Methods().ByName(updateMethod.Name())
	f.P(opts.response, ", ", opts.err, " ", opts.assign, " a.Service.", updateMethod.Name(), "(")
	f.P("a.ctx(),")
	f.P("&", pkg, ".", update.Input().Name(), "{")
	if opts.msg != "" {
		f.P(r.resource.Singular.UpperCamelCase(), ": ", opts.msg, ",")
	} else {
		f.P(r.resource.Singular.UpperCamelCase(), ": (func() *", messageType(pkg, r.message), " {")
		if hasParent(r.resource) {
			f.P("msg := a.Update(", opts.parent, ")")
		} else {
			f.P("msg := a.Update()")
		}
		f.P("msg.Name = ", opts.name)
		f.P("return msg")
		f.P("})(),")
	}
	if hasField(update.Input(), "update_mask") && len(opts.updateMaskPaths) > 0 {
		f.P("UpdateMask: &", fieldmask, ".FieldMask{")
		f.P("Paths: []string{")
		for _, path := range opts.updateMaskPaths {
			f.P(path, ",")
		}
		f.P("},")
		f.P("},")
	}
	f.P("},")
	f.P(")")
	if opts.assertNoError {
		f.P(assert, ".NilError(t, ", opts.err, ")")
	}
}

type callListOpts struct {
	parent        string
	pageSize      string
	pageToken     string
	response      string
	err           string
	assertNoError bool
}

func (r *resourceGenerator) callList(f *codegen.File, opts callListOpts) {
	pkg := f.Import(goImportPath(r.service.ParentFile())...)
	assert := f.Import("gotest.tools/v3/assert")
	listMethod, ok := r.resource.Methods[aipreflect.MethodTypeList]
	if !ok {
		panic(fmt.Errorf("no list method for resource '%s'", r.resource.Message))
	}
	list := r.service.Methods().ByName(listMethod.Name())
	if opts.err == "" {
		opts.err = defaultErr
	}

	f.P(opts.response, ", ", opts.err, " := a.Service.", listMethod.Name(), "(")
	f.P("a.ctx(),")
	f.P("&", pkg, ".", list.Input().Name(), "{")
	if hasParent(r.resource) {
		f.P("Parent: ", opts.parent, ",")
	}
	if opts.pageSize != "" {
		f.P("PageSize: ", opts.pageSize, ",")
	}
	if opts.pageToken != "" {
		f.P("PageToken: ", opts.pageToken, ",")
	}
	f.P("},")
	f.P(")")
	if opts.assertNoError {
		f.P(assert, ".NilError(t, err)")
	}
}

type callSearchOpts struct {
	parent        string
	pageSize      string
	pageToken     string
	response      string
	err           string
	assertNoError bool
}

func (r *resourceGenerator) callSearch(f *codegen.File, opts callSearchOpts) {
	pkg := f.Import(goImportPath(r.service.ParentFile())...)
	assert := f.Import("gotest.tools/v3/assert")
	searchMethod, ok := r.resource.Methods[aipreflect.MethodTypeSearch]
	if !ok {
		panic(fmt.Errorf("no search method for resource '%s'", r.resource.Message))
	}
	search := r.service.Methods().ByName(searchMethod.Name())
	if opts.err == "" {
		opts.err = defaultErr
	}

	f.P(opts.response, ", ", opts.err, " := a.Service.", searchMethod.Name(), "(")
	f.P("a.ctx(),")
	f.P("&", pkg, ".", search.Input().Name(), "{")
	if hasParent(r.resource) {
		f.P("Parent: ", opts.parent, ",")
	}
	if opts.pageSize != "" {
		f.P("PageSize: ", opts.pageSize, ",")
	}
	if opts.pageToken != "" {
		f.P("PageToken: ", opts.pageToken, ",")
	}
	f.P("},")
	f.P(")")
	if opts.assertNoError {
		f.P(assert, ".NilError(t, err)")
	}
}
