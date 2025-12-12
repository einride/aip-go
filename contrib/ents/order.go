package ents

import (
	"entgo.io/ent/dialect/sql"
	"go.einride.tech/aip/ordering"
)

// ApplyOrderBy builds a SQL order function from an orderBy string.
// Example: foo,bar asc/desc
func ApplyOrderBy(orderBy ordering.OrderBy) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		for _, field := range orderBy.Fields {
			if field.Desc {
				s.OrderBy(sql.Desc(s.C(field.Path)))
			} else {
				s.OrderBy(sql.Asc(s.C(field.Path)))
			}
		}
	}
}
