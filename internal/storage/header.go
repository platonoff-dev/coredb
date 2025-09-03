package storage

import (
	"encoding/binary"

	dberrors "github.com/platonoff-dev/coredb/pkg/corekv/errors"
)

const (
	HeaderSize = 4096
)

type DBHeader struct {
	Magic          []byte // Magic number to identify the database file
	Version        int    // Version of the database format
	PageSize       int    // Size of each page in the database
	FreeListPageID int    // Page ID of the freelist page
	PageCount      int    // Total number of pages in the database
}

func (h *DBHeader) Encode() []byte {
	data := make([]byte, HeaderSize)
	offset := 0

	copy(data[offset:], h.Magic)
	offset += len(h.Magic)

	n := binary.PutVarint(data[offset:], int64(h.Version))
	offset += n

	n = binary.PutVarint(data[offset:], int64(h.PageSize))
	offset += n

	n = binary.PutVarint(data[offset:], int64(h.FreeListPageID))
	offset += n

	_ = binary.PutVarint(data[offset:], int64(h.PageCount))

	return data
}

func (h *DBHeader) Decode(data []byte) error {
	if len(data) < HeaderSize {
		return dberrors.ErrInvalidFileFormat
	}

	offset := 0

	h.Magic = data[offset : offset+6]
	offset += 6

	version, n := binary.Varint(data[offset:])
	offset += n
	h.Version = int(version)

	pageSize, n := binary.Varint(data[offset:])
	offset += n
	h.PageSize = int(pageSize)

	freeListPageID, n := binary.Varint(data[offset:])
	offset += n
	h.FreeListPageID = int(freeListPageID)

	pageCount, _ := binary.Varint(data[offset:])
	h.PageCount = int(pageCount)

	return nil
}

func (h *DBHeader) IsValid() bool {
	// Check if the magic number is valid
	if len(h.Magic) != 6 {
		return false
	}
	// Compare only the non-zero bytes of magic (COREDB should be at the beginning)
	magicStr := string(h.Magic)
	if len(magicStr) < 6 || magicStr[:6] != "COREDB" {
		return false
	}
	// Check if the version is supported
	if h.Version != 1 {
		return false
	}
	// Check if the page size is a reasonable value
	if h.PageSize < 512 || h.PageSize > 65536 {
		return false
	}

	return true
}
