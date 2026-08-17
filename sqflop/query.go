package sqflop

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/lann/builder"
	"github.com/ysomad/flop"
	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/internal/ident"
)

// Cursor applies the resolved order, optional seek predicate, and over-fetch
// limit to base. Existing base ordering is rejected at ToSql time.
func Cursor(base sq.SelectBuilder, params flop.CursorParams) sq.SelectBuilder {
	if hasOrderBy(base) {
		return wrapErr(base, "%w", ErrBaseOrderBy)
	}
	orderBy := params.OrderBy()
	if len(orderBy.Fields) == 0 {
		return wrapErr(base, "%w", ErrNoOrder)
	}
	ordered, err := applyOrder(base, orderBy)
	if err != nil {
		return wrapErr(base, "%w", err)
	}
	after := params.After()
	if len(after) != 0 {
		predicate, err := cursorPredicate(orderBy, after)
		if err != nil {
			return wrapErr(ordered, "%w", err)
		}
		ordered = ordered.Where(predicate)
	}
	return ordered.Limit(uint64(params.Limit()))
}

// Page applies resolved ordering, LIMIT, and OFFSET to base.
func Page(base sq.SelectBuilder, params flop.OffsetParams) sq.SelectBuilder {
	if hasOrderBy(base) {
		return wrapErr(base, "%w", ErrBaseOrderBy)
	}
	if params.Limit() <= 0 || params.Offset() < 0 {
		return wrapErr(base, "invalid offset parameters")
	}
	orderBy := params.OrderBy()
	if len(orderBy.Fields) == 0 {
		return wrapErr(base, "%w", ErrNoOrder)
	}
	ordered, err := applyOrder(base, orderBy)
	if err != nil {
		return wrapErr(base, "%w", err)
	}
	ordered = ordered.Limit(uint64(params.Limit()))
	if params.Offset() > 0 {
		ordered = ordered.Offset(uint64(params.Offset()))
	}
	return ordered
}

// Count replaces selected columns with count(*) while preserving base filters.
func Count(base sq.SelectBuilder) sq.SelectBuilder {
	count := base.RemoveColumns().Column("count(*)").RemoveLimit().RemoveOffset()
	if hasOrderBy(base) {
		return wrapErr(count, "%w", ErrBaseOrderBy)
	}
	return count
}

func applyOrder(base sq.SelectBuilder, orderBy flop.OrderBy) (sq.SelectBuilder, error) {
	ordered := base
	for _, field := range orderBy.Fields {
		if !ident.ValidPath(field.Path) {
			return base, fmt.Errorf("invalid order column %q", field.Path)
		}
		ordered = ordered.OrderBy(field.SQL())
	}
	return ordered, nil
}

func cursorPredicate(orderBy flop.OrderBy, position flop.CursorPosition) (sq.Sqlizer, error) {
	if len(position) != len(orderBy.Fields) {
		return nil, ErrCursorPosition
	}
	disjunction := make(sq.Or, 0, len(position))
	equalities := make(sq.And, 0, len(position))
	for i, cursorField := range position {
		orderField := orderBy.Fields[i]
		if cursorField.Path != orderField.Path || cursorField.Value.Kind() == expr.NullKind {
			return nil, ErrCursorPosition
		}
		value := cursorField.Value.Any()
		var comparison sq.Sqlizer = sq.Gt{cursorField.Path: value}
		if orderField.Desc {
			comparison = sq.Lt{cursorField.Path: value}
		}
		if len(equalities) == 0 {
			disjunction = append(disjunction, comparison)
		} else {
			conjunction := make(sq.And, 0, len(equalities)+1)
			conjunction = append(conjunction, equalities...)
			conjunction = append(conjunction, comparison)
			disjunction = append(disjunction, conjunction)
		}
		equalities = append(equalities, sq.Eq{cursorField.Path: value})
	}
	if len(disjunction) == 1 {
		return disjunction[0], nil
	}
	return disjunction, nil
}

func hasOrderBy(base sq.SelectBuilder) bool {
	value, ok := builder.Get(base, "OrderByParts")
	if !ok {
		return false
	}
	parts, ok := value.([]sq.Sqlizer)
	return ok && len(parts) != 0
}
