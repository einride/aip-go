package aiptest

import (
	"strconv"

	"go.einride.tech/aip/internal/codegen"
	"go.einride.tech/aip/reflect/aipreflect"
)

func (r *resourceGenerator) getCase() testCase {
	_, ok := r.resource.Methods[aipreflect.MethodTypeGet]
	if !ok {
		return disabledTestCase()
	}
	return newTestCase("Get", func(f *codegen.File) {
		testing := f.Import("testing")
		assert := f.Import("gotest.tools/v3/assert")
		protocmp := f.Import("google.golang.org/protobuf/testing/protocmp")
		codes := f.Import("google.golang.org/grpc/codes")
		status := f.Import("google.golang.org/grpc/status")

		f.P("// Standard methods: Get")
		f.P("// https://google.aip.dev/131")
		f.P()

		if hasParent(r.resource) {
			f.P("parent := a.nextParent(t, false)")
		}
		r.callCreate(f, callCreateOpts{
			response:      "created00",
			parent:        "parent",
			assertNoError: true,
		})
		f.P()

		f.P("t.Run(\"missing\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callGet(f, callGetOpts{
			response: "_",
			err:      "err",
			name:     "created00.Name + \"notfound\"",
		})
		f.P(assert, ".Equal(t, ", codes, ".NotFound, ", status, ".Code(err), err)")
		f.P("})")
		f.P()

		f.P("t.Run(\"exists\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callGet(f, callGetOpts{
			response:      "msg",
			assertNoError: true,
			name:          "created00.Name",
		})
		f.P(assert, ".DeepEqual(t, created00, msg, ", protocmp, ".Transform())")
		f.P("})")
		f.P()

		f.P("t.Run(\"missing name\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callGet(f, callGetOpts{
			response: "_",
			err:      "err",
			name:     strconv.Quote(""),
		})
		f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
		f.P("})")
		f.P()

		f.P("t.Run(\"invalid name\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callGet(f, callGetOpts{
			response: "_",
			err:      "err",
			name:     strconv.Quote("invalid name"),
		})
		f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
		f.P("})")
		f.P()

		// TODO: add test for supplying wildcard as name
	})
}
