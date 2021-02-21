package aiptest

import (
	"fmt"
	"strconv"

	"go.einride.tech/aip/internal/codegen"
	"go.einride.tech/aip/reflect/aipreflect"
)

func (r *resourceGenerator) batchGetCase() testCase {
	_, ok := r.resource.Methods[aipreflect.MethodTypeBatchGet]
	if !ok {
		return disabledTestCase()
	}
	return newTestCase("BatchGet", func(f *codegen.File) {
		testing := f.Import("testing")
		pkg := f.Import(goImportPath(r.service.ParentFile())...)
		assert := f.Import("gotest.tools/v3/assert")
		protocmp := f.Import("google.golang.org/protobuf/testing/protocmp")
		codes := f.Import("google.golang.org/grpc/codes")
		status := f.Import("google.golang.org/grpc/status")

		f.P("// Batch methods: Get")
		f.P("// https://google.aip.dev/231")
		f.P()

		if hasParent(r.resource) {
			f.P("parent := a.nextParent(t, false)")
		}
		for i := 0; i < 3; i++ {
			r.callCreate(f, callCreateOpts{
				response:      fmt.Sprintf("created0%d", i),
				parent:        "parent",
				assertNoError: true,
			})
		}
		f.P()

		f.P("t.Run(\"all exists\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callBatchGet(f, callBatchGetOpts{
			parent:        "parent",
			names:         []string{"created00.Name", "created01.Name", "created02.Name"},
			response:      "response",
			assertNoError: true,
		})
		f.P(assert, ".DeepEqual(")
		f.P("t,")
		f.P("[]*", messageType(pkg, r.message), "{")
		f.P("created00,")
		f.P("created01,")
		f.P("created02,")
		f.P("},")
		f.P("response.", r.resource.Plural.UpperCamelCase(), ",")
		f.P(protocmp, ".Transform(),")
		f.P(")")
		f.P("})")
		f.P()

		f.P("// This test ensures that if caller supplies duplicate names, the service")
		f.P("// handles it correctly, ie. returns duplicate messages.")
		f.P("t.Run(\"duplicate names\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callBatchGet(f, callBatchGetOpts{
			parent:        "parent",
			names:         []string{"created00.Name", "created00.Name"},
			response:      "response",
			assertNoError: true,
		})
		f.P(assert, ".DeepEqual(")
		f.P("t,")
		f.P("[]*", messageType(pkg, r.message), "{")
		f.P("created00,")
		f.P("created00,")
		f.P("},")
		f.P("response.", r.resource.Plural.UpperCamelCase(), ",")
		f.P(protocmp, ".Transform(),")
		f.P(")")
		f.P("})")
		f.P()

		f.P("// The operation must be atomic: it must fail for all resources")
		f.P("// or succeed for all resources (no partial success). ")
		f.P("t.Run(\"one missing\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callBatchGet(f, callBatchGetOpts{
			parent:   "parent",
			names:    []string{"created00.Name", "created01.Name + \"notfound\"", "created02.Name"},
			response: "_",
			err:      "err",
		})
		f.P(assert, ".Equal(t, ", codes, ".NotFound, ", status, ".Code(err), err)")
		f.P("})")
		f.P()

		f.P("// If no resource names are provided, the API should error with INVALID_ARGUMENT.")
		f.P("t.Run(\"zero names\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callBatchGet(f, callBatchGetOpts{
			parent:   "parent",
			response: "_",
			err:      "err",
		})
		f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
		f.P("})")
		f.P()

		if hasParent(r.resource) {
			f.P("t.Run(\"invalid parent\", func(t *", testing, ".T) {")
			f.P("a.maybeSkip(t)")
			r.callBatchGet(f, callBatchGetOpts{
				parent:   strconv.Quote("invalid parent"),
				names:    []string{"created00.Name"},
				response: "_",
				err:      "err",
			})
			f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
			f.P("})")
			f.P()
		}

		f.P("t.Run(\"invalid name\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callBatchGet(f, callBatchGetOpts{
			parent:   "parent",
			names:    []string{strconv.Quote("invalid name")},
			response: "_",
			err:      "err",
		})
		f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
		f.P("})")
		f.P()

		if hasParent(r.resource) {
			f.P("// If a caller sets the \"parent\", and the parent collection in the name of any resource")
			f.P("// being retrieved does not match, the request must fail.")
			f.P("t.Run(\"parent mismatch\", func(t *", testing, ".T) {")
			f.P("a.maybeSkip(t)")
			r.callBatchGet(f, callBatchGetOpts{
				parent:   "a.peekNextParent(t)",
				names:    []string{"created00.Name"},
				response: "_",
				err:      "err",
			})
			f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
			f.P("})")
			f.P()
		}

		// TODO: add test for supplying wildcard in one of the names
	})
}
