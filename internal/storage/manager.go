package storage

import (
	dberrors "github.com/platonoff-dev/coredb/internal/errors"
)

type DBFileOperator interface {
	Close() error
	ReadAt(b []byte, off int64) (n int, err error)
	WriteAt(b []byte, off int64) (n int, err error)
	Truncate(size int64) error
}

// FilePageManager manager provides abstraction to work with disk.
type FilePageManager struct {
	Header   *DBHeader
	File     DBFileOperator
	PageSize uint32
}

func (pm *FilePageManager) Read(pageID uint32) (*RawPage, error) {
	if pageID == 0 {
		return nil, dberrors.ErrInvalidPageID
	}

	data := make([]byte, pm.PageSize)
	_, err := pm.File.ReadAt(data, int64(pageID*pm.PageSize))
	if err != nil {
		return nil, err
	}

	page := &RawPage{}
	page.Decode(pageID, data)

	return page, nil
}

func (pm *FilePageManager) Write(page *RawPage) error {
	if page == nil {
		return dberrors.ErrInvalidPageID
	}

	data := page.Encode(pm.PageSize)
	_, err := pm.File.WriteAt(data, int64(page.ID*pm.PageSize))
	return err
}

func (pm *FilePageManager) Allocate() (*RawPage, error) {
	page := &RawPage{
		ID:              pm.Header.PageCount + 1,
		Type:            PageTypeBTreeLeaf,
		FreeSpaceOffset: 0,
		Data:            make([]byte, pm.PageSize),
	}

	err := pm.File.Truncate(int64(pm.Header.PageCount+1) * int64(pm.PageSize))
	if err != nil {
		return nil, err
	}

	pm.Header.PageCount++
	return page, nil
}

func (pm *FilePageManager) Free(pageID uint32) error {
	if pageID == 0 {
		return dberrors.ErrInvalidPageID
	}

	_, err := pm.Read(pageID)
	if err != nil {
		return err
	}

	return nil
}
