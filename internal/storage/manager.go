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
	File     DBFileOperator
	Header   DBHeader
	PageSize uint32
}

func (m *FilePageManager) Read(pageID uint32) (*RawPage, error) {
	if pageID == 0 {
		return nil, dberrors.ErrInvalidPageID
	}

	if m.File == nil {
		return nil, dberrors.ErrInvalidFileFormat
	}

	data := make([]byte, m.PageSize)
	_, err := m.File.ReadAt(data, int64(pageID*m.PageSize))
	if err != nil {
		return nil, err
	}

	page := &RawPage{}
	page.Decode(pageID, data)

	return page, nil
}

func (m *FilePageManager) Write(page *RawPage) error {
	if page == nil {
		return dberrors.ErrInvalidPageID
	}

	if m.File == nil {
		return dberrors.ErrInvalidFileFormat
	}

	data := page.Encode(m.PageSize)
	_, err := m.File.WriteAt(data, int64(page.ID*m.PageSize))
	return err
}

func (m *FilePageManager) Allocate() (*RawPage, error) {
	if m.File == nil {
		return nil, dberrors.ErrInvalidFileFormat
	}

	page := &RawPage{
		ID:              m.Header.PageCount + 1,
		Type:            PageTypeBTreeLeaf,
		FreeSpaceOffset: 0,
		Data:            make([]byte, m.PageSize),
	}

	err := m.File.Truncate(int64(m.Header.PageCount+1) * int64(m.PageSize))
	if err != nil {
		return nil, err
	}

	err = m.Write(page)
	if err != nil {
		return nil, err
	}

	m.Header.PageCount++

	return page, nil
}

func (m *FilePageManager) Free(pageID uint32) error {
	if pageID == 0 {
		return dberrors.ErrInvalidPageID
	}

	if m.File == nil {
		return dberrors.ErrInvalidFileFormat
	}

	_, err := m.Read(pageID)
	if err != nil {
		return err
	}

	return nil
}
