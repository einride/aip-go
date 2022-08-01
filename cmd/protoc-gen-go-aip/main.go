package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go.einride.tech/aip/cmd/protoc-gen-go-aip/internal/genaip"
	"google.golang.org/protobuf/compiler/protogen"
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
		includeServiceName = flags.Bool(
			"include_service_name",
			false,
			"when generating resource definitions, set to true to include the service name in the struct name",
		)
		removeServiceNameSuffix = flags.String(
			"remove_service_name_suffix",
			"",
			"when including the resource definition service name, remove this suffix before generating the struct name."+
				"Can be a '+' separated list (eg. 'domain.com+domain.net')",
		)
	)
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		return genaip.Run(plugin, genaip.Config{
			IncludeResourceDefinitions: *includeResourceDefinitions,
			IncludeServiceName:         *includeServiceName,
			RemoveServiceNameSuffix:    strings.Split(*removeServiceNameSuffix, "+"),
		})
	})
}
