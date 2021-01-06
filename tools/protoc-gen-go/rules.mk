protobuf_go_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
protoc_gen_go := $(protobuf_go_cwd)/bin/protoc-gen-go
export PATH := $(dir $(protoc_gen_go)):$(PATH)

$(protoc_gen_go): $(protobuf_go_cwd)/../../go.mod
	$(info [protoc-gen-go] building binary...)
	@cd $(protobuf_go_cwd)/../.. && go build -o $@ google.golang.org/protobuf/cmd/protoc-gen-go
