package storage

import "encoding/binary"

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
