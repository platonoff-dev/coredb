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
	return nil, dberrors.ErrNotImplemented
}

func (pm *FilePageManager) Write(page *RawPage) error {
	return dberrors.ErrNotImplemented
}

func (pm *FilePageManager) Allocate() (*RawPage, error) {
	return nil, dberrors.ErrNotImplemented
}

func (pm *FilePageManager) Free() error {
	return dberrors.ErrNotImplemented
}
