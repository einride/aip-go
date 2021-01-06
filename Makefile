SHELL := /bin/bash

all: \
	commitlint \
	prettier-markdown \
	buf-lint \
	buf-generate \
	go-lint \
	go-review \
	go-test \
	go-mod-tidy \
	git-verify-nodiff

include tools/buf/rules.mk
include tools/commitlint/rules.mk
include tools/git-verify-nodiff/rules.mk
include tools/golangci-lint/rules.mk
include tools/goreview/rules.mk
include tools/prettier/rules.mk
include tools/protoc-gen-go/rules.mk
include tools/protoc/rules.mk
include tools/semantic-release/rules.mk

.PHONY: examples/proto/api-common-protos
examples/proto/api-common-protos:
	@git submodule update --init --recursive $@

.PHONY: go-mod-tidy
go-mod-tidy:
	$(info [$@] tidying Go module files...)
	@go mod tidy -v

.PHONY: go-test
go-test:
	$(info [$@] running Go tests...)
	@go test -count 1 -cover -race ./...

.PHONY: buf-lint
buf-lint: $(buf) examples/proto/api-common-protos
	$(info [$@] linting protobuf schemas...)
	@$(buf) lint

.PHONY: buf-generate
buf-generate: $(buf) $(protoc) $(protoc_gen_go) examples/proto/api-common-protos
	$(info [$@] generating protobuf stubs...)
	@rm -rf examples/proto/gen
	@$(buf) generate --path examples/proto/src/einride
