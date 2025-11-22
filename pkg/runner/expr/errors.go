package expr

import "fmt"

// Error is an error type caused by lexing/parsing expression syntax. For more details, see
// https://docs.github.com/en/actions/learn-github-actions/expressions
type Error struct {
	// Message is an error message
	Message string
	// Offset is byte offset position which caused the error
	Offset int
	// Offset is line number position which caused the error. Note that this value is 1-based.
	Line int
	// Column is column number position which caused the error. Note that this value is 1-based.
	Column int
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d:%d:%d: %s", e.Line, e.Column, e.Offset, e.Message)
}

func (e *Error) String() string {
	return e.Error()
}
