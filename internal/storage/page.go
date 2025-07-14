// Package storage provides functionality for managing database storage layer.
package storage

import (
	dberrors "github.com/platonoff-dev/coredb/internal/errors"
)

const (
	PageTypeDBHeader byte = iota + 1
	PageTypeFreelist
	PageTypeBTreeInternal
	PageTypeBTreeLeaf
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

func (rp *RawPage) AsFreelist() (*FreelistPage, error) {
	return nil, dberrors.ErrNotImplemented
}
