package rules

import "go.einride.tech/aip/lint"

func All() []lint.Rule {
	return []lint.Rule{
		&AIP4231ResourceReferenceAnnotation{},
	}
}
