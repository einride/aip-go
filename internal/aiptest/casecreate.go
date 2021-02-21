package aiptest

import (
	"strconv"

	"go.einride.tech/aip/internal/codegen"
	"go.einride.tech/aip/reflect/aipreflect"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (r *resourceGenerator) createCase() testCase {
	createMethod, ok := r.resource.Methods[aipreflect.MethodTypeCreate]
	if !ok {
		return disabledTestCase()
	}
	create := r.service.Methods().ByName(createMethod.Name())
	_, ok = r.resource.Methods[aipreflect.MethodTypeGet]
	if !ok {
		return disabledTestCase()
	}

	return newTestCase("Create", func(f *codegen.File) {
		testing := f.Import("testing")
		assert := f.Import("gotest.tools/v3/assert")
		protocmp := f.Import("google.golang.org/protobuf/testing/protocmp")
		codes := f.Import("google.golang.org/grpc/codes")
		status := f.Import("google.golang.org/grpc/status")
		time := f.Import("time")
		protoReflect := f.Import("google.golang.org/protobuf/reflect/protoreflect")
		stringsPkg := f.Import("strings")

		f.P("// Standard methods: Create")
		f.P("// https://google.aip.dev/133")
		f.P()

		f.P("_ = protoreflect.ValueOf")
		f.P()

		if hasParent(r.resource) {
			f.P("parent := a.nextParent(t, false)")
			f.P()
		}

		f.P("t.Run(\"persisted\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callCreate(f, callCreateOpts{
			response:      "msg",
			parent:        "parent",
			assertNoError: true,
		})
		r.callGet(f, callGetOpts{
			response:      "persisted",
			name:          "msg.Name",
			assertNoError: true,
		})
		f.P(assert, ".DeepEqual(t, msg, persisted, ", protocmp, ".Transform())")
		f.P("})")
		f.P()

		f.P("t.Run(\"create time\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callCreate(f, callCreateOpts{
			response:      "msg",
			parent:        "parent",
			assertNoError: true,
		})
		f.P(assert, ".Check(t, ", time, ".Since(msg.CreateTime.AsTime()) < ", time, ".Second)")
		f.P("})")
		f.P()

		if hasRequiredFields(r.message) {
			f.P("t.Run(\"required fields\", func(t *", testing, ".T) {")
			f.P("a.maybeSkip(t)")
			f.P("for _, tt := range []", protoReflect, ".Name{")
			rangeRequiredFields(r.message, func(field protoreflect.FieldDescriptor) {
				f.P(strconv.Quote(string(field.Name())), ",")
			})
			f.P("} {")
			f.P("tt := tt")
			f.P("t.Run(string(tt), func(t *", testing, ".T) {")
			if hasParent(r.resource) {
				f.P("msg := a.Create(parent)")
			} else {
				f.P("msg := a.Create()")
			}
			f.P("msg.ProtoReflect().Clear(msg.ProtoReflect().Descriptor().Fields().ByName(tt))")
			r.callCreate(f, callCreateOpts{
				parent:   "parent",
				msg:      "msg",
				err:      "err",
				response: "_",
			})
			f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
			f.P("})")
			f.P("}")
			f.P("})")
			f.P()
		}

		if hasParent(r.resource) {
			f.P("t.Run(\"missing parent\", func(t *", testing, ".T) {")
			f.P("a.maybeSkip(t)")
			r.callCreate(f, callCreateOpts{
				parent:   strconv.Quote(""),
				response: "_",
				msg:      "a.Create(parent)",
				err:      "err",
			})
			f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
			f.P("})")
			f.P()

			f.P("t.Run(\"invalid parent\", func(t *", testing, ".T) {")
			f.P("a.maybeSkip(t)")
			r.callCreate(f, callCreateOpts{
				parent:   strconv.Quote("invalid parent"),
				response: "_",
				msg:      "a.Create(parent)",
				err:      "err",
			})
			f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
			f.P("})")
			f.P()
		}

		if hasUserSettableID(r.resource, create) {
			f.P("t.Run(\"user settable id\", func(t *", testing, ".T) {")
			f.P("a.maybeSkip(t)")
			r.callCreate(f, callCreateOpts{
				parent:        "parent",
				response:      "msg",
				assertNoError: true,
				id:            strconv.Quote("usersetid"),
			})
			f.P(assert, ".Check(t, ", stringsPkg, ".HasSuffix(msg.Name, \"usersetid\"))")
			f.P("})")
			f.P()

			f.P("t.Run(\"already exists\", func(t *", testing, ".T) {")
			f.P("a.maybeSkip(t)")
			r.callCreate(f, callCreateOpts{
				parent:        "parent",
				response:      "_",
				assertNoError: true,
				id:            strconv.Quote("alreadyexists"),
			})
			r.callCreate(f, callCreateOpts{
				parent:   "parent",
				response: "_",
				err:      "err",
				assign:   "=",
				id:       strconv.Quote("alreadyexists"),
			})
			f.P(assert, ".Equal(t, ", codes, ".AlreadyExists, ", status, ".Code(err), err)")
			f.P("})")
			f.P()
		}
		// TODO: add test for supplying wildcard as parent.
	})
}
