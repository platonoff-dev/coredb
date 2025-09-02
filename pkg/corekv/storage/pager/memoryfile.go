package pager

import (
	"io"
	"os"
)

type MemoryFile struct {
	Data []byte
}

func (f *MemoryFile) Close() error {
	// No-op for memory file
	return nil
}

func (f *MemoryFile) ReadAt(b []byte, off int64) (n int, err error) {
	if off < 0 || int(off) >= len(f.Data) {
		return 0, os.ErrInvalid
	}
	n = copy(b, f.Data[off:])
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

func (f *MemoryFile) WriteAt(b []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, os.ErrInvalid
	}
	if int(off)+len(b) > len(f.Data) {
		newData := make([]byte, int(off)+len(b))
		copy(newData, f.Data)
		f.Data = newData
	}
	n = copy(f.Data[off:], b)
	return n, nil
}

func (f *MemoryFile) Truncate(size int64) error {
	if size < 0 {
		return os.ErrInvalid
	}
	if int(size) < len(f.Data) {
		f.Data = f.Data[:size]
	} else {
		f.Data = append(f.Data, make([]byte, int(size)-len(f.Data))...)
	}
	return nil
}
