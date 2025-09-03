package pager

import (
	"errors"
	"io"
)

var (
	ErrPageNotExist  = errors.New("page not exist")
	ErrPageInvalid   = errors.New("page invalid")
	ErrDiskOperation = errors.New("disk operation")
)

type DbFile interface {
	io.ReaderAt
	io.WriterAt
	Truncate(size int64) error
}

// PageManager manager provides abstraction to work with disk.
type PageManager struct {
	File DbFile

	PageSize  int
	PageCount int
}

func (m *PageManager) Read(pageID int) ([]byte, error) {
	if pageID < 0 || pageID >= m.PageCount {
		return nil, ErrPageNotExist
	}

	data := make([]byte, m.PageSize)
	offset := pageID * m.PageSize
	_, err := m.File.ReadAt(data, int64(offset))
	if err != nil {
		return nil, errors.Join(ErrDiskOperation, err)
	}

	return data, nil
}

func (m *PageManager) Write(id int, data []byte) error {
	if id < 0 || id >= m.PageCount {
		return ErrPageNotExist
	}

	if data == nil || len(data) != m.PageSize {
		return ErrPageInvalid
	}

	_, err := m.File.WriteAt(data, int64(id*m.PageSize))
	if err != nil {
		return errors.Join(ErrDiskOperation, err)
	}

	return nil
}

func (m *PageManager) Allocate() (int, error) {
	m.PageCount++

	err := m.File.Truncate(int64(m.PageCount * m.PageSize))
	if err != nil {
		return 0, err
	}

	return m.PageCount - 1, nil
}
