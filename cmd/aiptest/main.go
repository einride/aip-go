package main

import (
	"flag"
	"log"
	"os"

	"go.einride.tech/aip/internal/aiptest"
	"gopkg.in/yaml.v2"
)

func main() {
	log.SetFlags(0)
	configFilePath := flag.String("config", "aip-test.yaml", "config file")
	flag.Parse()
	if flag.Arg(0) != "generate" {
		log.Fatal("usage: aip-test generate [-config <config>]")
	}
	configFile, err := os.Open(*configFilePath)
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		if err := configFile.Close(); err != nil {
			log.Panic(err)
		}
	}()
	var config aiptest.Config
	if err := yaml.NewDecoder(configFile).Decode(&config); err != nil {
		log.Panic(err)
	}
	for _, pkg := range config.Packages {
		if err := aiptest.GeneratePackage(pkg); err != nil {
			log.Panic(err)
		}
	}
}
