// Package apperr defines the error kinds that services return.
// Handlers translate these into HTTP status codes (see httpx.Error),
// so services never need to know about HTTP.
package apperr

import (
	"errors"
	"fmt"
)

// Kinds. Check them with errors.Is(err, apperr.ErrNotFound).
var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
	ErrInvalid  = errors.New("invalid input")
)

// Error carries a user-facing message plus one of the kinds above.
type Error struct {
	Kind    error
	Message string
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Kind }

func NotFound(format string, args ...any) error {
	return &Error{Kind: ErrNotFound, Message: fmt.Sprintf(format, args...)}
}

func Conflict(format string, args ...any) error {
	return &Error{Kind: ErrConflict, Message: fmt.Sprintf(format, args...)}
}

func Invalid(format string, args ...any) error {
	return &Error{Kind: ErrInvalid, Message: fmt.Sprintf(format, args...)}
}
