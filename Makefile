SHELL := /bin/bash

all: \
	commitlint \
	prettier-markdown \
	proto \
	go-stringer \
	buf-generate-testdata \
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
include tools/semantic-release/rules.mk
include tools/stringer/rules.mk

.PHONY: proto
proto:
	$(info [$@] building protos...)
	@make -C proto

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

.PHONY: buf-generate-testdata
buf-generate-testdata: $(buf) build/protoc-gen-go-aip
	$(info [$@] generating testdata stubs...)
	@cd cmd/protoc-gen-go-aip/internal/genaip/testdata && $(buf) generate --path test
