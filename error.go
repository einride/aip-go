package flop

import (
	"errors"
	"fmt"
)

// Sentinels for every failure flop reports. Each names what flop found wrong,
// not what a caller should do about it: mapping a cause onto a status code, a
// retry, or a log level is the caller's decision, and flop has no view on it.
//
//	params, err := paginator.CursorParams(cursor, size, orderBy, filter)
//	switch {
//	case errors.Is(err, flop.ErrInvalidCursor),
//		errors.Is(err, flop.ErrInvalidOrder),
//		errors.Is(err, flop.ErrInvalidFilter),
//		errors.Is(err, flop.ErrInvalidPageSize):
//		// the values that reached flop were unusable
//	case err != nil:
//		// ErrDeclaration: this endpoint is misdeclared
//	}
var (
	// ErrInvalidFilter reports a filter flop refuses: one that does not parse,
	// names an undeclared field, carries a value of the wrong type, uses an
	// operator the field does not accept, exceeds the declared FilterLimits, or
	// arrives at an endpoint where filtering is disabled. A syntax failure is
	// also a *FilterError, which pins it to a line and column.
	ErrInvalidFilter = errors.New("flop: invalid filter")

	// ErrInvalidCursor reports a cursor that is malformed, truncated, or
	// carries a value that cannot address a row.
	ErrInvalidCursor = errors.New("flop: invalid cursor")

	// ErrCursorMismatch reports a well-formed cursor issued for a different
	// order or filter than the one accompanying it.
	ErrCursorMismatch = errors.New("flop: cursor does not match")

	// ErrInvalidOrder reports an order naming an undeclared field, repeating a
	// field, or using a malformed path.
	ErrInvalidOrder = errors.New("flop: invalid order")

	// ErrInvalidPageSize reports a negative page size.
	ErrInvalidPageSize = errors.New("flop: invalid page size")

	// ErrInvalidPage reports a negative page number.
	ErrInvalidPage = errors.New("flop: invalid page number")

	// ErrDeclaration reports an endpoint declared wrongly. Constructors return
	// it and no input can provoke it, so it always marks a programmer mistake.
	ErrDeclaration = errors.New("flop: invalid declaration")

	// ErrTooManyRows reports more rows handed to a page constructor than the
	// parameters asked for.
	ErrTooManyRows = errors.New("flop: too many rows")
)

// FilterError is a filter syntax error, reported at the position where it
// occurred. It matches ErrInvalidFilter.
type FilterError struct {
	Pos   FilterPosition
	Msg   string
	cause error
}

func (e *FilterError) Error() string {
	return fmt.Sprintf("flop: filter: %d:%d: %s", e.Pos.Line, e.Pos.Column, e.Msg)
}

// Unwrap reports the classifying sentinel alongside any underlying cause, so
// errors.Is finds both.
func (e *FilterError) Unwrap() []error {
	if e.cause == nil {
		return []error{ErrInvalidFilter}
	}
	return []error{ErrInvalidFilter, e.cause}
}

// errorf builds an error that leads with sentinel and details it with format,
// rendering as "flop: invalid order: unknown field \"total\"".
func errorf(sentinel error, format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{sentinel}, args...)...)
}
