semantic_release_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
semantic_release := $(semantic_release_cwd)/node_modules/.bin/semantic-release

$(semantic_release): $(semantic_release_cwd)/package.json
	$(info [semantic-release] installing packages...)
	@cd $(semantic_release_cwd) && npm install --no-save --no-audit --ignore-scripts &> /dev/null
	@touch $@

.PHONY: semantic-release
semantic-release: $(semantic_release_cwd)/.releaserc.yaml $(semantic_release)
	$(info [$@] creating release...)
	@cd $(semantic_release_cwd) && $(semantic_release)
