golangci_lint_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
golangci_lint_version := 1.41.1
golangci_lint := $(golangci_lint_cwd)/$(golangci_lint_version)/golangci-lint

ifeq ($(shell uname),Linux)
golangci_lint_archive_url := https://github.com/golangci/golangci-lint/releases/download/v${golangci_lint_version}/golangci-lint-${golangci_lint_version}-linux-amd64.tar.gz
else ifeq ($(shell uname),Darwin)
golangci_lint_archive_url := https://github.com/golangci/golangci-lint/releases/download/v${golangci_lint_version}/golangci-lint-${golangci_lint_version}-darwin-amd64.tar.gz
else
$(error unsupported OS: $(shell uname))
endif

$(golangci_lint):
	$(info [golangci-lint] fetching version $(golangci_lint) binary...)
	@mkdir -p $(dir $@)
	@curl -sSL $(golangci_lint_archive_url) -o - | \
		tar -xz --directory $(dir $@) --strip-components 1
	@chmod +x $@
	@touch $@

.PHONY: go-lint
go-lint: $(golangci_lint)
	$(info [$@] linting Go code...)
	@$(golangci_lint) run
