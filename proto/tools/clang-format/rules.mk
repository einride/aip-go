clang_format_dir := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
clang_format_version := 1.4.0
clang_format := $(clang_format_dir)/bin/clang-format
export PATH := $(dir $(clang_format)):$(PATH)

ifeq ($(shell uname),Linux)
clang_format_bin_path := $(clang_format_dir)/node_modules/clang-format/bin/linux_x64
else ifeq ($(shell uname),Darwin)
clang_format_bin_path := $(clang_format_dir)/node_modules/clang-format/bin/darwin_x64
else
$(error unsupported OS: $(shell uname))
endif

$(clang_format):
	$(info [clang-format] installing version $(clang_format_version)...)
	@npm install --no-save --no-audit --prefix $(clang_format_dir) clang-format@$(clang_format_version) &> /dev/null
	@mkdir -p $(dir $(clang_format))
	@ln -fs $(clang_format_bin_path)/clang-format $@
	@chmod +x $@
	@touch $@
