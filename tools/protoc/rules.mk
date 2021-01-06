protoc_cwd := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
protoc_version := 3.14.0
protoc_dir := $(protoc_cwd)/$(protoc_version)
protoc := $(protoc_dir)/bin/protoc
export PATH := $(dir $(protoc)):$(PATH)

ifeq ($(shell uname),Linux)
protoc_zip_url := https://github.com/protocolbuffers/protobuf/releases/download/v$(protoc_version)/protoc-$(protoc_version)-linux-$(shell uname -m).zip
else ifeq ($(shell uname),Darwin)
protoc_zip_url := https://github.com/protocolbuffers/protobuf/releases/download/v$(protoc_version)/protoc-$(protoc_version)-osx-$(shell uname -m).zip
else
$(error unsupported OS: $(shell uname))
endif

$(protoc):
	$(info [protoc] fetching version $(protoc_version)...)
	@mkdir -p $(protoc_dir)
	@curl -sSLo $(protoc_dir)/archive.zip $(protoc_zip_url)
	@unzip -d $(protoc_dir) $(protoc_dir)/archive.zip
	@chmod +x $@
	@touch $@
