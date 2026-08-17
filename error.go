package flop

import (
	"fmt"
	"strconv"
)

// FilterError is a filter syntax error, reported at the position where it
// occurred.
type FilterError struct {
	Pos FilterPosition
	Msg string
	err error
}

func newFilterError(pos FilterPosition, wrapped error, format string, args ...any) *FilterError {
	return &FilterError{Pos: pos, Msg: fmt.Sprintf(format, args...), err: wrapped}
}

func (e *FilterError) Error() string {
	return "flop: filter: " +
		strconv.FormatInt(int64(e.Pos.Line), 10) + ":" +
		strconv.FormatInt(int64(e.Pos.Column), 10) + ": " + e.Msg
}

func (e *FilterError) Unwrap() error { return e.err }
