mage_dir := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
mage_tools_path := $(mage_dir)/tools
mage := $(mage_tools_path)/mgmake/magefile

$(mage): $(mage_dir)/go.mod $(mage_dir)/*.go
	@cd $(mage_dir) && go run go.einride.tech/mage-tools gen

.PHONY: mage-clean
mage-clean:
	@git clean -fdx $(mage_dir)
