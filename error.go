package flop

import "fmt"

// FilterError is a filter syntax error, reported at the position where it
// occurred.
type FilterError struct {
	Pos   FilterPosition
	Msg   string
	cause error
}

func (e *FilterError) Error() string {
	return fmt.Sprintf("flop: filter: %d:%d: %s", e.Pos.Line, e.Pos.Column, e.Msg)
}

func (e *FilterError) Unwrap() error { return e.cause }
