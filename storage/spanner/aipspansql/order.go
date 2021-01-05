package aipspansql

import (
	"strings"

	"cloud.google.com/go/spanner/spansql"
	"go.einride.tech/aip/ordering"
)

// Order translates a valid ordering.OrderBy expression to a spansql.Order expression.
func Order(orderBy ordering.OrderBy) []spansql.Order {
	if len(orderBy.Fields) == 0 {
		return nil
	}
	result := make([]spansql.Order, 0, len(orderBy.Fields))
	for _, field := range orderBy.Fields {
		subFields := strings.Split(field.Path, ".")
		if len(subFields) == 1 {
			result = append(result, spansql.Order{Expr: spansql.ID(subFields[0]), Desc: field.Desc})
			continue
		}
		pathExp := make(spansql.PathExp, 0, len(subFields))
		for _, subField := range subFields {
			pathExp = append(pathExp, spansql.ID(subField))
		}
		result = append(result, spansql.Order{Expr: pathExp, Desc: field.Desc})
	}
	return result
}
