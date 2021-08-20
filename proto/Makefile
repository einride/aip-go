SHELL := /usr/bin/env bash

.PHONY: all
all: \
	clang-format \
	buf-lint \
	api-linter-lint \
	buf-generate

include tools/api-linter/rules.mk
include tools/buf/rules.mk
include tools/clang-format/rules.mk
include tools/protoc-gen-go-grpc/rules.mk

build/protoc-gen-go: ../go.mod
	$(info [$@] building binary...)
	@go build -o $@ google.golang.org/protobuf/cmd/protoc-gen-go

.PHONY: build/protoc-gen-go-aip
build/protoc-gen-go-aip:
	$(info [$@] building binary...)
	@go build -o $@ ../cmd/protoc-gen-go-aip

build/descriptor.pb: $(shell find einride -type f -name '*.proto')
	$(info [$@] generating proto descriptor...)
	@mkdir -p $(dir $@)
	@$(buf) build -o $@

.PHONY: buf-lint
buf-lint: $(buf)
	$(info [$@] linting proto schemas...)
	@$(buf) lint

.PHONY: clang-format
clang-format: proto_files = $(shell find einride -type f -name '*.proto')
clang-format: $(clang_format)
	$(info [$@] formatting proto files...)
	@$(clang_format) \
		-i \
		--style='{BasedOnStyle: Google, ColumnLimit: 0, Language: Proto}' \
		$(proto_files)

.PHONY: api-linter-lint
api-linter-lint: $(api_linter) build/descriptor.pb
	$(info [$@] linting gRPC APIs...)
	@$(api_linter) \
		--set-exit-status \
		--config api-linter.yaml \
		--descriptor-set-in build/descriptor.pb \
		$(shell find einride/example/freight -type f -name '*.proto')

protoc_plugins := \
	$(protoc_gen_go_grpc) \
	build/protoc-gen-go \
	build/protoc-gen-go-aip

.PHONY: buf-generate
buf-generate: $(buf) $(protoc_plugins)
	$(info [$@] generating protobuf stubs...)
	@$(buf) generate \
		--template buf.gen.yaml \
		--path einride
