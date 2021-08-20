api_linter_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
api_linter := $(api_linter_cwd)/bin/api-linter
export PATH := $(dir $(api_linter)):$(PATH)

api_linter_version := 1.26.2
api_linter_zip_url := https://github.com/googleapis/api-linter/releases/download/v$(api_linter_version)/api-linter-$(api_linter_version)-$(shell uname -s)-amd64.tar.gz

$(api_linter): $(api_linter_cwd)/rules.mk
	$(info [api-linter] fetching $(api_linter_version) binary...)
	@mkdir -p $(dir $@)
	@curl -sSL $(api_linter_zip_url) -o - | tar -xz --directory $(dir $@)
	@touch $@
