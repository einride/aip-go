package sqflop

import (
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

var (
	ErrBaseOrderBy    = errors.New("base query has an ORDER BY")
	ErrNoOrder        = errors.New("pagination has no resolved order")
	ErrCursorPosition = errors.New("cursor position does not match the resolved order")
)

// Err returns a Sqlizer that carries an error through a chainable builder.
func Err(err error) sq.Sqlizer { return errSqlizer{err: err} }

type errSqlizer struct{ err error }

func (e errSqlizer) ToSql() (string, []any, error) { return "", nil, e.err }

func wrapErr(base sq.SelectBuilder, format string, args ...any) sq.SelectBuilder {
	return base.Where(Err(fmt.Errorf("sqflop: "+format, args...)))
}
