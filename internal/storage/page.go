// Package storage package implements logic needed for basic operations with disk, B-Tree, etc
package storage

import (
	dberrors "github.com/platonoff-dev/coredb/internal/errors"
)

const (
	PageTypeDBHeader byte = iota + 1
)

type RawPage struct {
	Data            []byte
	ID              uint32
	FreeSpaceOffset uint32
	Type            byte
}

type FreelistPage struct {
	FreePages    []uint32
	NextFreePage uint32
}

type SchemaPage struct{}

func (rp *RawPage) AsFreelist() (*FreelistPage, error) {
	return nil, dberrors.ErrNotImplemented
}
