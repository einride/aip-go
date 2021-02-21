package aiptest

import (
	"strconv"

	"go.einride.tech/aip/internal/codegen"
	"go.einride.tech/aip/reflect/aipreflect"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (r *resourceGenerator) updateCase() testCase {
	updateMethod, ok := r.resource.Methods[aipreflect.MethodTypeUpdate]
	if !ok {
		return disabledTestCase()
	}
	update := r.service.Methods().ByName(updateMethod.Name())
	_ = update
	_, ok = r.resource.Methods[aipreflect.MethodTypeGet]
	if !ok {
		return disabledTestCase()
	}
	_, ok = r.resource.Methods[aipreflect.MethodTypeCreate]
	if !ok {
		return disabledTestCase()
	}

	return newTestCase("Update", func(f *codegen.File) {
		testing := f.Import("testing")
		assert := f.Import("gotest.tools/v3/assert")
		protocmp := f.Import("google.golang.org/protobuf/testing/protocmp")
		codes := f.Import("google.golang.org/grpc/codes")
		status := f.Import("google.golang.org/grpc/status")
		protoReflect := f.Import("google.golang.org/protobuf/reflect/protoreflect")

		f.P("// Standard methods: Update")
		f.P("// https://google.aip.dev/134")
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
			parent:        "parent",
			response:      "created",
			assertNoError: true,
		})
		r.callUpdate(f, callUpdateOpts{
			parent:        "parent",
			name:          "created.Name",
			response:      "updated",
			assertNoError: true,
		})
		r.callGet(f, callGetOpts{
			response:      "persisted",
			name:          "updated.Name",
			assertNoError: true,
		})
		f.P(assert, ".DeepEqual(t, updated, persisted, ", protocmp, ".Transform())")
		f.P("})")
		f.P()

		f.P("t.Run(\"update time\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callCreate(f, callCreateOpts{
			response:      "created",
			parent:        "parent",
			assertNoError: true,
		})
		r.callUpdate(f, callUpdateOpts{
			parent:        "parent",
			name:          "created.Name",
			response:      "updated",
			assertNoError: true,
		})
		f.P(assert, ".Check(t, updated.UpdateTime.AsTime().After(created.UpdateTime.AsTime()))")
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
			r.callCreate(f, callCreateOpts{
				response:      "created",
				parent:        "parent",
				assertNoError: true,
			})
			if hasParent(r.resource) {
				f.P("msg := a.Update(parent)")
			} else {
				f.P("msg := a.Update()")
			}
			f.P("msg.Name = created.Name")
			f.P("msg.ProtoReflect().Clear(msg.ProtoReflect().Descriptor().Fields().ByName(tt))")
			r.callUpdate(f, callUpdateOpts{
				msg:             "msg",
				err:             "err",
				updateMaskPaths: []string{strconv.Quote("*")},
				assign:          "=",
				response:        "_",
			})
			f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
			f.P("})")
			f.P("}")
			f.P("})")
			f.P()
		}

		if hasField(update.Input(), "update_mask") {
			f.P("t.Run(\"invalid update mask\", func(t *", testing, ".T) {")
			f.P("a.maybeSkip(t)")
			r.callCreate(f, callCreateOpts{
				response:      "created",
				parent:        "parent",
				assertNoError: true,
			})
			r.callUpdate(f, callUpdateOpts{
				parent:          "parent",
				name:            "created.Name",
				updateMaskPaths: []string{strconv.Quote("invalid_field_xxx")},
				response:        "_",
				err:             "err",
				assign:          "=",
			})
			f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
			f.P("})")
			f.P()
		}

		f.P("t.Run(\"missing name\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callUpdate(f, callUpdateOpts{
			parent:   "parent",
			name:     strconv.Quote(""),
			response: "_",
			err:      "err",
		})
		f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
		f.P("})")
		f.P()

		f.P("t.Run(\"invalid name\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callUpdate(f, callUpdateOpts{
			parent:   "parent",
			name:     strconv.Quote(""),
			response: "_",
			err:      "err",
		})
		f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
		f.P("})")
		f.P()

		// TODO: add test for supplying wildcard as parent.
		// TODO: add test for etags
	})
}
