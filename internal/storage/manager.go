package storage

import "errors"

var (
	ErrPageNotFound  = errors.New("page not exist")
	ErrPageInvalid   = errors.New("page invalid")
	ErrDiskOperation = errors.New("disk operation")
)

type DBFileOperator interface {
	Close() error
	ReadAt(b []byte, off int64) (n int, err error)
	WriteAt(b []byte, off int64) (n int, err error)
	Truncate(size int64) error
}

// FilePageManager manager provides abstraction to work with disk.
type FilePageManager struct {
	File DBFileOperator

	PageSize  int
	PageCount int
}

func (m *FilePageManager) Read(pageID int) ([]byte, error) {
	if pageID == 0 {
		return nil, ErrPageNotFound
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
	if data == nil || len(data) != m.PageSize {
		return ErrPageInvalid
	}

	_, err := m.File.WriteAt(data, int64(id*m.PageSize))
	return err
}

func (m *FilePageManager) Allocate() (int, error) {
	err := m.File.Truncate(int64((m.PageCount + 1) * m.PageSize))
	if err != nil {
		return 0, err
	}

	m.PageCount++

	return m.PageCount - 1, nil
}

func (m *FilePageManager) Free(id int) error {
	_, err := m.Read(id)
	if err != nil {
		return err
	}

	// TODO: Actually free pages when freelist will be implementedz

	return nil
}
