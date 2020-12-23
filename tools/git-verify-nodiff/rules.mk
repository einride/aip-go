git_verify_nodiff_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
git_verify_nodiff := $(git_verify_nodiff_cwd)/git-verify-nodiff.bash

.PHONY: git-verify-nodiff
git-verify-nodiff:
	$(info [$@] verifying that git has no diff...)
	@$(git_verify_nodiff)
