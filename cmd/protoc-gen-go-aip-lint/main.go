package main

import (
	"log"

	"go.einride.tech/aip/lint"
	"go.einride.tech/aip/lint/rules"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	log.SetFlags(0)
	linter, err := lint.NewLinter(rules.All()...)
	if err != nil {
		log.Fatal(err)
	}
	protogen.Options{ParamFunc: linter.ParamFunc}.Run(linter.Run)
}
