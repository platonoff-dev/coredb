package errors

import "errors"

var (
	ErrNotImplemented    = errors.New("not implemented")
	ErrInvalidFileFormat = errors.New("invalid file format")
	ErrInvalidPageID     = errors.New("invalid page ID")
)
