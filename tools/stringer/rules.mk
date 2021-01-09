stringer_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
stringer := $(stringer_cwd)/bin/stringer
export PATH := $(PATH):$(dir $(stringer))

$(stringer): $(stringer_cwd)/go.mod
	$(info [stringer] building binary...)
	@cd $(stringer_cwd) && go build -o $@ golang.org/x/tools/cmd/stringer
	@cd $(stringer_cwd) && go mod tidy
