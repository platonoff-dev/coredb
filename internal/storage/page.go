// Package storage package implements logic needed for basic operations with disk, B-Tree, etc
package storage

import (
	"encoding/binary"
	"errors"
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
	if rp.Type != PageTypeDBHeader {
		return nil, errors.New("not a freelist page")
	}

	freePages := make([]uint32, 0)
	for i := len(rp.Data); i > 8; i -= 4 {
		pageID := binary.LittleEndian.Uint32(rp.Data[i : i+4])
		if pageID == 0 {
			break
		}
		freePages = append(freePages, pageID)
	}

	nextFreePage := binary.LittleEndian.Uint32(rp.Data[0:4])

	return &FreelistPage{
		FreePages:    freePages,
		NextFreePage: nextFreePage,
	}, nil
}
