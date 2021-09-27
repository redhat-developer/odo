package locale

import (
	"errors"
)

var (
	// ErrNotDetected returns while no locale detected.
	ErrNotDetected = errors.New("not detected")
	// ErrNotSupported means current platform or language is not supported.
	ErrNotSupported = errors.New("not supported")
)

// Error is the error returned by locale.
type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	return e.Op + ": " + e.Err.Error()
}

// Unwrap implements xerrors.Wrapper
func (e *Error) Unwrap() error {
	return e.Err
}
