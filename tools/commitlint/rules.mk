commitlint_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
commitlint := $(commitlint_cwd)/node_modules/.bin/commitlint

$(commitlint): $(commitlint_cwd)/package.json
	$(info [commitlint] installing package...)
	@cd $(commitlint_cwd) && npm install --no-save --no-audit &> /dev/null
	@touch $@

.PHONY: commitlint
commitlint: $(commitlint_cwd)/.commitlintrc.js $(commitlint)
	$(info [$@] linting commit messages...)
	@git fetch --tags
	@NODE_PATH=$(commitlint_cwd)/node_modules $(commitlint) \
		--config $< \
		--from origin/master \
		--to HEAD
