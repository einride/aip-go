package main

import (
	"log"
	"os"
	"path/filepath"

	"go.einride.tech/aip/cmd/protoc-gen-go-aip/internal/genaip"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		log.Printf("%v %v\n", filepath.Base(os.Args[0]), genaip.PluginVersion)
		os.Exit(0)
	}
	protogen.Options{}.Run(genaip.Run)
}
