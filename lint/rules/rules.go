package rules

import (
	"go.einride.tech/aip/lint"
	"go.einride.tech/aip/lint/rules/aip4231"
)

func All() []lint.Rule {
	return []lint.Rule{
		&aip4231.ResourceReferenceAnnotation{},
	}
}
