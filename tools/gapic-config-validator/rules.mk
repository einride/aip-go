gapic_config_validator_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
gapic_config_validator_version := 0.6.0
protoc_gen_gapic_validator := $(gapic_config_validator_cwd)/bin/protoc-gen-gapic-validator
PATH := $(dir $(protoc_gen_gapic_validator)):$(PATH)

protoc_gen_gapic_validator_zip_url := https://github.com/googleapis/gapic-config-validator/releases/download/v$(gapic_config_validator_version)/gapic-config-validator-$(gapic_config_validator_version)-$(shell uname -s)-amd64.tar.gz

$(protoc_gen_gapic_validator): $(gapic_config_validator_cwd)/rules.mk
	$(info [gapic-config-validator] fetching version $(gapic_config_validator_version)...)
	@mkdir -p $(dir $@)
	@curl -sSL $(protoc_gen_gapic_validator_zip_url) -o - | tar -xz --directory $(dir $@)
	@touch $@
