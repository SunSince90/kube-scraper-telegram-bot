package backend

import "errors"

var (
	// ErrNotFound is returned when the resource was not found
	ErrNotFound error = errors.New("not found")
)
