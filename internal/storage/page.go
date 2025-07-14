// Package storage package implements logic needed for basic operations with disk, B-Tree, etc
package storage

import (
	"encoding/binary"
	"errors"
	"os"
)

const (
	PageTypeDBHeader byte = iota + 1
)

type RawPage struct {
	ID              uint64
	Type            byte
	FreeSpaceOffset uint32
	Data            []byte
}

type FreelistPage struct {
	FreePages    []uint64
	NextFreePage uint64
}

type SchemaPage struct {
}

func (rp *RawPage) AsFreelist() (*FreelistPage, error) {
	if rp.Type != PageTypeDBHeader {
		return nil, errors.New("not a freelist page")
	}

	freePages := make([]uint64, 0)
	for i := len(rp.Data); i > 8; i -= 8 {
		pageID := binary.LittleEndian.Uint64(rp.Data[i : i+8])
		if pageID == 0 {
			break
		}
		freePages = append(freePages, pageID)
	}

	nextFreePage := binary.LittleEndian.Uint64(rp.Data[0:8])

	return &FreelistPage{
		FreePages:    freePages,
		NextFreePage: nextFreePage,
	}, nil
}

// PageManager manager provides abstraction to work with disk.
type PageManager struct {
	file     *os.File
	pageSize uint64
}

func (pm *PageManager) FetchPage(pageID uint64) (*RawPage, error) {
	buff := make([]byte, pm.pageSize)
	offset := int64(pageID * pm.pageSize)
	_, err := pm.file.ReadAt(buff, offset)
	if err != nil {
		return nil, err
	}

	return &RawPage{
		ID:              pageID,
		Type:            buff[0],
		FreeSpaceOffset: binary.LittleEndian.Uint32(buff[1:5]),
		Data:            buff[5:],
	}, nil
}

func (pm *PageManager) WritePage(page *RawPage) error {
	buff := make([]byte, pm.pageSize)
	buff[0] = page.Type
	binary.LittleEndian.PutUint32(buff[1:5], page.FreeSpaceOffset)
	copy(buff[5:], page.Data)

	offset := int64(page.ID * pm.pageSize)
	_, err := pm.file.WriteAt(buff, offset)
	if err != nil {
		return err
	}

	return nil
}
