package storage

import (
	"encoding/binary"
	"errors"
	"os"
)

// PageManager manager provides abstraction to work with disk.
type PageManager struct {
	file     *os.File
	pageSize uint32
}

func (pm *PageManager) Read(pageID uint32) (*RawPage, error) {
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

func (pm *PageManager) Write(page *RawPage) error {
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

func (pm *PageManager) Allocate() (*RawPage, error) {
	return nil, errors.New("not implemented")
}

func (pm *PageManager) Free() error {
	return errors.New("not implemented")
}
