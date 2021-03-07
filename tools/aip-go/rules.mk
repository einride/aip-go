aip_go_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
protoc_gen_go_aip_lint := $(aip_go_cwd)/bin/protoc-gen-go-aip-lint
export PATH := $(dir $(protoc_gen_go_aip_lint)):$(PATH)

.PHONY: $(protoc_gen_go_aip_lint)
$(protoc_gen_go_aip_lint):
	$(info [protoc-gen-go-aip-lint] building binary...)
	@cd $(aip_go_cwd)/../.. && go build -o $@ ./cmd/protoc-gen-go-aip-lint
