SHELL := /bin/bash

all: \
	commitlint \
	prettier-markdown \
	go-stringer \
	buf-lint \
	buf-generate \
	buf-generate-testdata \
	go-lint \
	go-review \
	go-test \
	go-mod-tidy \
	git-verify-nodiff

include tools/api-linter/rules.mk
include tools/buf/rules.mk
include tools/commitlint/rules.mk
include tools/git-verify-nodiff/rules.mk
include tools/golangci-lint/rules.mk
include tools/goreview/rules.mk
include tools/prettier/rules.mk
include tools/protoc-gen-go-grpc/rules.mk
include tools/semantic-release/rules.mk
include tools/stringer/rules.mk

build/protoc-gen-go: go.mod
	$(info [$@] rebuilding plugin...)
	@go build -o $@ google.golang.org/protobuf/cmd/protoc-gen-go

.PHONY: build/protoc-gen-go-aip
build/protoc-gen-go-aip:
	$(info [$@] rebuilding plugin...)
	@go build -o $@ ./cmd/protoc-gen-go-aip

.PHONY: go-stringer
go-stringer: \
	reflect/aipreflect/methodtype_string.go

%_string.go: %.go $(stringer)
	$(info [stringer] generating $@ from $<)
	@go generate ./$<

.PHONY: go-test
go-test:
	$(info [$@] running Go tests...)
	@go test -count 1 -cover -race ./...

.PHONY: go-mod-tidy
go-mod-tidy:
	$(info [$@] tidying Go module files...)
	@go mod tidy -v

.PHONY: proto/gen/descriptor.pb
proto/gen/descriptor.pb: $(buf)
	$(info [$@] generating proto descriptor...)
	@mkdir -p $(dir $@)
	@$(buf) build -o $@

.PHONY: buf-lint
buf-lint: $(buf)
	$(info [$@] linting protobuf schemas...)
	@$(buf) lint

.PHONY: api-linter-lint
api-linter-lint: $(api_linter_wrapper) proto/gen/descriptor.pb
	$(info [$@] linting APIs...)
	@$(api_linter_wrapper) \
		--config api-linter.yaml \
		--descriptor-set-in proto/gen/descriptor.pb \
		-I proto/src \
		$(shell find proto/src -type f -name '*.proto' | cut -d '/' -f 4-)

protoc_plugins := \
	build/protoc-gen-go \
	build/protoc-gen-go-aip \
	$(protoc_gen_go_grpc)

.PHONY: buf-generate
buf-generate: $(buf) $(protoc) $(protoc_plugins)
	$(info [$@] generating protobuf stubs...)
	@rm -rf proto/gen
	@$(buf) generate --path proto/src/einride

.PHONY: buf-generate-testdata
buf-generate-testdata: build/protoc-gen-go-aip
	$(info [$@] generating testdata stubs...)
	@cd cmd/protoc-gen-go-aip/internal/genaip/testdata && $(buf) generate --path test
