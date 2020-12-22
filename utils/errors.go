package utils

import "errors"

// predefined errors
var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNotSupport     = errors.New("not support")
	ErrNullPointer    = errors.New("null pointer")
	ErrInvalidArgs    = errors.New("invalid arguments")
)
