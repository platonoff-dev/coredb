package heap

import (
	"encoding/binary"
)

const (
	HeaderSize = 1 + binary.MaxVarintLen64*3
)

type Page struct {
	RecordMap       map[int64][]uint64 // Row ID to Record Offset mapping and segment length
	Data            []byte
	ID              int64
	FreeSpaceOffset uint64
	WritableSpace   uint64
	NextPageID      int64
	Type            byte
}

func (p *Page) MarshalBinary() ([]byte, error) {
	data := []byte{}
	data = append(data, p.Type)
	binary.PutUvarint(data, p.FreeSpaceOffset)
	binary.PutUvarint(data, p.WritableSpace)
	binary.PutVarint(data, p.NextPageID)
	data = append(data, p.Data...)

	return data, nil
}

func (p *Page) UnmarshalBinary(id int64, data []byte) error {
	p.ID = id

	p.Type = data[0]
	ptr := 1

	freeSpaceOffset, n := binary.Uvarint(data[ptr:])
	p.FreeSpaceOffset = freeSpaceOffset
	ptr += n

	p.WritableSpace, n = binary.Uvarint(data[ptr:])
	ptr += n

	nextPageID, _ := binary.Varint(data[ptr:])
	p.NextPageID = nextPageID

	p.Data = data[ptr:]
	return nil
}
