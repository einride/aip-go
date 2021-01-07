protoc_gen_go_grpc_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
protoc_gen_go_grpc := $(protoc_gen_go_grpc_cwd)/bin/protoc-gen-go-grpc
export PATH := $(dir $(protoc_gen_go_grpc)):$(PATH)

$(protoc_gen_go_grpc): $(protoc_gen_go_grpc_cwd)/go.mod
	$(info [protoc-gen-go-grpc] building binary...)
	@cd $(protoc_gen_go_grpc_cwd) && go build -o $@ google.golang.org/grpc/cmd/protoc-gen-go-grpc
	@cd $(protoc_gen_go_grpc_cwd) && go mod tidy -v
