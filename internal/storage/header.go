package storage

import (
	"encoding/binary"

	dberrors "github.com/platonoff-dev/coredb/internal/errors"
)

type DBHeader struct {
	Magic          []byte // Magic number to identify the database file
	Version        uint16 // Version of the database format
	PageSize       uint32 // Size of each page in the database
	FreeListPageID uint32 // Page ID of the freelist page
}

func (h *DBHeader) Encode() []byte {
	// Encode the header into a byte slice
	data := make([]byte, 0)
	data = append(data, h.Magic...)
	binary.LittleEndian.AppendUint16(data, h.Version)
	binary.LittleEndian.AppendUint32(data, h.PageSize)
	binary.LittleEndian.AppendUint32(data, h.FreeListPageID)
	return data
}

func (h *DBHeader) Decode(data []byte) error {
	if len(data) < 12 {
		return dberrors.ErrInvalidFileFormat
	}

	h.Magic = data[:8]
	h.Version = binary.LittleEndian.Uint16(data[8:10])
	h.PageSize = binary.LittleEndian.Uint32(data[10:14])
	h.FreeListPageID = binary.LittleEndian.Uint32(data[14:18])

	return nil
}

func (h *DBHeader) IsValid() bool {
	// Check if the magic number is valid
	if len(h.Magic) != 8 || string(h.Magic) != "COREDB" {
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
