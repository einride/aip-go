aip_go_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
protoc_gen_go_aip := $(aip_go_cwd)/bin/protoc-gen-go-aip
export PATH := $(dir $(protoc_gen_go_aip)):$(PATH)

.PHONY: $(protoc_gen_go_aip)
$(protoc_gen_go_aip):
	$(info [protoc-gen-go-aip] building binary...)
	@cd $(aip_go_cwd)/../.. && go build -o $@ ./cmd/protoc-gen-go-aip
