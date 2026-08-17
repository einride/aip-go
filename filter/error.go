package filter

import (
	"fmt"
	"strconv"
)

// Error is a filter syntax error, reported at the position where it occurred.
type Error struct {
	Pos Position
	Msg string
	err error
}

func newError(pos Position, wrapped error, format string, args ...any) *Error {
	return &Error{Pos: pos, Msg: fmt.Sprintf(format, args...), err: wrapped}
}

func (e *Error) Error() string {
	return "filter: " +
		strconv.FormatInt(int64(e.Pos.Line), 10) + ":" +
		strconv.FormatInt(int64(e.Pos.Column), 10) + ": " + e.Msg
}

func (e *Error) Unwrap() error { return e.err }
