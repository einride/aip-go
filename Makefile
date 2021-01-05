SHELL := /bin/bash

all: \
	commitlint \
	prettier-markdown \
	go-lint \
	go-review \
	go-test \
	go-mod-tidy \
	git-verify-nodiff

include tools/commitlint/rules.mk
include tools/git-verify-nodiff/rules.mk
include tools/golangci-lint/rules.mk
include tools/goreview/rules.mk
include tools/prettier/rules.mk
include tools/semantic-release/rules.mk

go_module_dirs := \
	. \
	storage/spanner

.PHONY: go-mod-tidy
go-mod-tidy:
	$(info [$@] tidying Go module files...)
	@for dir in $(go_module_dirs); do \
		cd $$dir && go mod tidy -v; \
	done

.PHONY: go-test
go-test:
	$(info [$@] running Go tests...)
	@for dir in $(go_module_dirs); do \
		cd $$dir && go test -count 1 -cover -race ./...; \
	done
