package pager

import (
	dberrors "github.com/platonoff-dev/coredb/pkg/corekv/errors"
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
	PageSize int
}

func (m *FilePageManager) Read(pageID int) ([]byte, error) {
	if pageID == 0 {
		return nil, dberrors.ErrInvalidPageID
	}

	if m.File == nil {
		return nil, dberrors.ErrInvalidFileFormat
	}

	data := make([]byte, m.PageSize)
	offset := pageID * m.PageSize
	_, err := m.File.ReadAt(data, int64(offset))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (m *FilePageManager) Write(id int, data []byte) error {
	if m.File == nil {
		return dberrors.ErrInvalidFileFormat
	}

	if data == nil || len(data) != m.PageSize {
		return dberrors.ErrInvalidPageFormat
	}

	_, err := m.File.WriteAt(data, int64(id*m.PageSize))
	return err
}

func (m *FilePageManager) Allocate() (int, error) {
	if m.File == nil {
		return 0, dberrors.ErrInvalidFileFormat
	}

	err := m.File.Truncate(int64((m.Header.PageCount + 1) * m.PageSize))
	if err != nil {
		return 0, err
	}

	m.Header.PageCount++

	return m.Header.PageCount - 1, nil
}

func (m *FilePageManager) Free(id int) error {
	if id == 0 {
		return dberrors.ErrInvalidPageID
	}

	if m.File == nil {
		return dberrors.ErrInvalidFileFormat
	}

	_, err := m.Read(id)
	if err != nil {
		return err
	}

	// TODO: Actually free pages when freelist will be implemented

	return nil
}
