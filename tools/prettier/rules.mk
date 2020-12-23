prettier_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
prettier := $(prettier_cwd)/node_modules/.bin/prettier

$(prettier): $(prettier_cwd)/package.json
	$(info [prettier] installing...)
	@cd $(prettier_cwd) && npm install --no-save --no-audit &> /dev/null
	@touch $@

.PHONY: prettier-markdown
prettier-markdown: $(prettier_cwd)/.prettierignore $(prettier)
	$(info [$@] formatting Markdown files...)
	@$(prettier) \
		--loglevel warn \
		--ignore-path $< \
		--parser markdown \
		--prose-wrap always \
		--write **/*.md
