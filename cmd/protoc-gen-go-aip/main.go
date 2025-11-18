package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"go.einride.tech/aip/cmd/protoc-gen-go-aip/internal/genaip"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		log.Printf("%v %v\n", filepath.Base(os.Args[0]), genaip.PluginVersion)
		os.Exit(0)
	}
	var (
		flags                      flag.FlagSet
		includeResourceDefinitions = flags.Bool(
			"include_resource_definitions",
			true,
			"set to false to exclude resource definitions from code generation",
		)
	)
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		return genaip.Run(plugin, genaip.Config{
			IncludeResourceDefinitions: *includeResourceDefinitions,
		})
	})
}
