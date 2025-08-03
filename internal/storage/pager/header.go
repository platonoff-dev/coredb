package pager

import (
	"encoding/binary"

	dberrors "github.com/platonoff-dev/coredb/internal/errors"
)

type DBHeader struct {
	Magic          []byte // Magic number to identify the database file
	Version        uint16 // Version of the database format
	PageSize       uint32 // Size of each page in the database
	FreeListPageID uint32 // Page ID of the freelist page
	PageCount      uint32 // Total number of pages in the database
}

func (h *DBHeader) Encode() []byte {
	data := []byte{}

	data = append(data, h.Magic...)
	data = binary.LittleEndian.AppendUint16(data, h.Version)
	data = binary.LittleEndian.AppendUint32(data, h.PageSize)
	data = binary.LittleEndian.AppendUint32(data, h.FreeListPageID)
	data = binary.LittleEndian.AppendUint32(data, h.PageCount)
	return data
}

func (h *DBHeader) Decode(data []byte) error {
	if len(data) < 20 {
		return dberrors.ErrInvalidFileFormat
	}

	h.Magic = data[:6]
	h.Version = binary.LittleEndian.Uint16(data[6:8])
	h.PageSize = binary.LittleEndian.Uint32(data[8:12])
	h.FreeListPageID = binary.LittleEndian.Uint32(data[12:16])
	h.PageCount = binary.LittleEndian.Uint32(data[16:20])

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
