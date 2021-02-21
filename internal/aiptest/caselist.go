package aiptest

import (
	"strconv"

	"go.einride.tech/aip/internal/codegen"
	"go.einride.tech/aip/reflect/aipreflect"
)

func (r *resourceGenerator) listCase() testCase {
	_, ok := r.resource.Methods[aipreflect.MethodTypeList]
	if !ok {
		return disabledTestCase()
	}
	_, ok = r.resource.Methods[aipreflect.MethodTypeCreate]
	if !ok {
		return disabledTestCase()
	}

	return newTestCase("List", func(f *codegen.File) {
		testing := f.Import("testing")
		pkg := f.Import(goImportPath(r.service.ParentFile())...)
		assert := f.Import("gotest.tools/v3/assert")
		protocmp := f.Import("google.golang.org/protobuf/testing/protocmp")
		cmpopts := f.Import("github.com/google/go-cmp/cmp/cmpopts")
		codes := f.Import("google.golang.org/grpc/codes")
		status := f.Import("google.golang.org/grpc/status")

		f.P("// Standard methods: List")
		f.P("// https://google.aip.dev/132")
		f.P()

		if hasParent(r.resource) {
			f.P("parent01 := a.nextParent(t, false)")
			f.P("parent02 := a.nextParent(t, true)")
			f.P()
		} else {
			f.P("_ = ", cmpopts, ".SortSlices")
		}

		// create 15 under each parent
		f.P("const n = 15")
		f.P()
		if hasParent(r.resource) {
			f.P("parent01msgs := make([]*", messageType(pkg, r.message), ", n)")
			f.P("for i := 0; i < n; i++ {")
			r.callCreate(f, callCreateOpts{
				response:      "msg",
				parent:        "parent01",
				assertNoError: true,
			})
			f.P("parent01msgs[i] = msg")
			f.P("}")
			f.P()
		}
		f.P("parent02msgs := make([]*", messageType(pkg, r.message), ", n)")
		f.P("for i := 0; i < n; i++ {")
		r.callCreate(f, callCreateOpts{
			response:      "msg",
			parent:        "parent02",
			assertNoError: true,
		})
		f.P("parent02msgs[i] = msg")
		f.P("}")
		f.P()

		if hasParent(r.resource) {
			f.P("// list methods with a specified parent should not return resources")
			f.P("// owned by another parent.")
			f.P("t.Run(\"isolation\", func(t *", testing, ".T) {")
			f.P("a.maybeSkip(t)")
			r.callList(f, callListOpts{
				parent:        "parent02",
				pageSize:      "9999",
				response:      "response",
				assertNoError: true,
			})
			f.P(assert, ".DeepEqual(")
			f.P("t,")
			f.P("parent02msgs,")
			f.P("response.", r.resource.Plural.UpperCamelCase(), ",")
			f.P(protocmp, ".Transform(),")
			f.P(cmpopts, ".SortSlices(func(a,b *", messageType(pkg, r.message), ") bool {")
			f.P("return a.Name < b.Name")
			f.P("}),")
			f.P(")")
			f.P("})")
			f.P()

			f.P("t.Run(\"pagination\", func(t *", testing, ".T) {")
			f.P("a.maybeSkip(t)")

			f.P("// If there are no more resources, next_page_token should be unset.")
			f.P("t.Run(\"next page token\", func(t *", testing, ".T) {")
			f.P("a.maybeSkip(t)")
			r.callList(f, callListOpts{
				parent:        "parent02",
				pageSize:      "n",
				response:      "response",
				assertNoError: true,
			})
			f.P("assert.Equal(t, \"\", response.NextPageToken)")
			f.P("})")
			f.P()

			f.P("// Listing resource one by one should eventually return all resources created.")
			f.P("// Catches errors where page tokens are not stable.")
			f.P("t.Run(\"one by one\", func(t *", testing, ".T) {")
			f.P("a.maybeSkip(t)")
			f.P("msgs := make([]*", messageType(pkg, r.message), ", 0, n)")
			f.P("var nextPageToken string")
			f.P("for {")
			r.callList(f, callListOpts{
				parent:        "parent02",
				pageSize:      "1",
				pageToken:     "nextPageToken",
				response:      "response",
				assertNoError: true,
			})
			f.P(assert, ".Equal(t, 1, len(response.", r.resource.Plural.UpperCamelCase(), "))")
			f.P("msgs = append(msgs, response.", r.resource.Plural.UpperCamelCase(), "...)")
			f.P("nextPageToken = response.NextPageToken")
			f.P("if nextPageToken == \"\" {")
			f.P("break")
			f.P("}")
			f.P("}")
			f.P(assert, ".DeepEqual(")
			f.P("t,")
			f.P("parent02msgs,")
			f.P("msgs,")
			f.P(protocmp, ".Transform(),")
			f.P(cmpopts, ".SortSlices(func(a,b *", messageType(pkg, r.message), ") bool {")
			f.P("return a.Name < b.Name")
			f.P("}),")
			f.P(")")
			f.P("})")
			f.P()

			f.P("})")
			f.P()

			if deleteMethod, ok := r.resource.Methods[aipreflect.MethodTypeDelete]; ok {
				del := r.service.Methods().ByName(deleteMethod.Name())
				f.P("t.Run(\"deleted\", func(t *", testing, ".T) {")
				f.P("a.maybeSkip(t)")
				f.P("// Delete some of the resources")
				f.P("const nDelete = 5")
				f.P("for i := 0; i < nDelete; i++ {")
				f.P("_, err := a.Service.", del.Name(), "(a.ctx(), &", pkg, ".", del.Input().Name(), "{")
				f.P("Name: parent02msgs[i].Name,")
				f.P("})")
				f.P(assert, ".NilError(t, err)")
				f.P("}")
				r.callList(f, callListOpts{
					parent:        "parent02",
					pageSize:      "9999",
					response:      "response",
					assertNoError: true,
				})
				f.P(assert, ".DeepEqual(")
				f.P("t,")
				f.P("parent02msgs[nDelete:],")
				f.P("response.", r.resource.Plural.UpperCamelCase(), ",")
				f.P(protocmp, ".Transform(),")
				f.P(cmpopts, ".SortSlices(func(a,b *", messageType(pkg, r.message), ") bool {")
				f.P("return a.Name < b.Name")
				f.P("}),")
				f.P(")")
				f.P("})")
			}

			f.P("t.Run(\"invalid parent\", func(t *", testing, ".T) {")
			f.P("a.maybeSkip(t)")
			r.callList(f, callListOpts{
				parent:   strconv.Quote("invalid parent"),
				response: "_",
				err:      "err",
			})
			f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
			f.P("})")
			f.P()
		}

		f.P("t.Run(\"negative page size\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callList(f, callListOpts{
			parent:   "parent02",
			pageSize: "-10",
			response: "_",
			err:      "err",
		})
		f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
		f.P("})")
		f.P()

		f.P("t.Run(\"invalid page token\", func(t *", testing, ".T) {")
		f.P("a.maybeSkip(t)")
		r.callList(f, callListOpts{
			parent:    "parent02",
			pageToken: strconv.Quote("invalid page token"),
			response:  "_",
			err:       "err",
		})
		f.P(assert, ".Equal(t, ", codes, ".InvalidArgument, ", status, ".Code(err), err)")
		f.P("})")
	})
}
