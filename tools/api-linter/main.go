package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	log.SetFlags(0)
	// nolint: gosec // OK to forward command line arguments here.
	cmd := exec.Command("api-linter", append([]string{"--output-format", "json"}, os.Args[1:]...)...)
	var stdoutBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	var output []struct {
		FilePath string `json:"file_path"`
		Problems []struct {
			Message    string `json:"message"`
			Suggestion string `json:"suggestion,omitempty"`
			Location   struct {
				Start struct {
					Line   int `json:"line_number"`
					Column int `json:"column_number"`
				} `json:"start_position"`
			} `json:"location"`
			RuleID     string `json:"rule_id"`
			RuleDocURI string `json:"rule_doc_uri"`
		} `json:"problems"`
	}
	if err := json.NewDecoder(&stdoutBuffer).Decode(&output); err != nil {
		log.Fatal(err)
	}
	var hasProblem bool
	for _, file := range output {
		for _, problem := range file.Problems {
			hasProblem = true
			fmt.Printf(
				"\n%s:%d:%d:\n\t%s %s\n\t%s\n\t%s\n",
				file.FilePath,
				problem.Location.Start.Line,
				problem.Location.Start.Column,
				problem.Message,
				problem.Suggestion,
				problem.RuleID,
				problem.RuleDocURI,
			)
		}
	}
	if hasProblem {
		os.Exit(1)
	}
}
