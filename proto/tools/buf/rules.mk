buf_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
buf_version := 0.51.1
buf := $(buf_cwd)/$(buf_version)/bin/buf
export PATH := $(dir $(buf)):$(PATH)

os := $(shell uname -s)-$(shell uname -m)

# enforce x86 architecture if mac m1 until tool has official support
ifeq ($(os),Darwin-arm64)
os = Darwin-x86_64
endif

buf_bin_url := https://github.com/bufbuild/buf/releases/download/v$(buf_version)/buf-$(os)

$(buf): $(buf_cwd)/rules.mk
	$(info [buf] feching $(buf_version) binary...)
	@mkdir -p $(dir $@)
	@curl -sSL $(buf_bin_url) -o $@
	@chmod +x $@
	@touch $@
