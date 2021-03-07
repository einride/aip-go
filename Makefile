SHELL := /bin/bash

all: \
	commitlint \
	prettier-markdown \
	buf-lint \
	buf-generate \
	go-generate \
	go-lint \
	go-review \
	go-test \
	go-mod-tidy \
	git-verify-nodiff

include tools/aip-go/rules.mk
include tools/api-linter/rules.mk
include tools/buf/rules.mk
include tools/commitlint/rules.mk
include tools/gapic-config-validator/rules.mk
include tools/git-verify-nodiff/rules.mk
include tools/golangci-lint/rules.mk
include tools/goreview/rules.mk
include tools/prettier/rules.mk
include tools/protoc-gen-go-grpc/rules.mk
include tools/protoc-gen-go/rules.mk
include tools/protoc/rules.mk
include tools/semantic-release/rules.mk
include tools/stringer/rules.mk

.PHONY: examples/proto/api-common-protos
examples/proto/api-common-protos:
	@git submodule update --init --recursive $@

.PHONY: go-generate
go-generate: \
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

.PHONY: api-linter-lint
api-linter-lint: $(api_linter_wrapper)
	$(info [$@] linting APIs...)
	@$(api_linter_wrapper) \
		--config api-linter.yaml \
		-I examples/proto/api-common-protos \
		-I examples/proto/src \
		$(shell find examples/proto/src -type f -name '*.proto' | cut -d '/' -f 4-)

.PHONY: buf-lint
buf-lint: $(buf) examples/proto/api-common-protos
	$(info [$@] linting protobuf schemas...)
	@$(buf) lint

protoc_plugins := \
	$(protoc_gen_go) \
	$(protoc_gen_go_grpc) \
	$(protoc_gen_gapic_validator) \
	$(protoc_gen_go_aip_lint)

.PHONY: buf-generate
buf-generate: $(buf) $(protoc) $(protoc_plugins) examples/proto/api-common-protos
	$(info [$@] generating protobuf stubs...)
	@rm -rf examples/proto/gen
	@$(buf) generate --path examples/proto/src/einride
