// Package storage provides functionality for managing database storage layer.
package storage

import (
	"encoding/binary"
)

const (
	PageTypeDBHeader byte = iota + 1
	PageTypeFreelist
	PageTypeBTreeInternal
	PageTypeBTreeLeaf
)

type RawPage struct {
	Data            []byte
	ID              uint32
	FreeSpaceOffset uint32
	Type            byte
}

func (p *RawPage) Encode(pageSize uint32) []byte {
	data := make([]byte, pageSize)

	data[0] = p.Type
	binary.LittleEndian.PutUint32(data[1:5], p.FreeSpaceOffset)
	copy(data[5:], p.Data)
	return data
}

func (p *RawPage) Decode(id uint32, data []byte) {
	p.ID = id
	p.Type = data[0]
	p.FreeSpaceOffset = binary.LittleEndian.Uint32(data[1:5])
	p.Data = data[5:]
}
