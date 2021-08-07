buf_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
buf_version := 0.48.2
buf := $(buf_cwd)/$(buf_version)/buf

buf_bin_url := https://github.com/bufbuild/buf/releases/download/v$(buf_version)/buf-$(shell uname -s)-$(shell uname -m)

$(buf): $(buf_cwd)/rules.mk
	$(info [buf] feching $(buf_version) binary...)
	@mkdir -p $(dir $@)
	@curl -sSL $(buf_bin_url) -o $@
	@chmod +x $@
	@touch $@
