package heap

import (
	"encoding/binary"
	"errors"
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

func (p *Page) getRecord(rowID int64) ([]byte, bool) {
	pointers, exists := p.RecordMap[rowID]
	if !exists {
		return nil, false
	}

	return p.Data[pointers[0] : pointers[0]+pointers[1]], true
}

// TODO:  When updating a record calclulation of frespace is little bit different.
// We need to find old record remove it then calculate space for new record and only if it still fits we can write it.
// TODO: Write test first to handle this case.
func (p *Page) setRecord(rowID int64, record []byte) error {
	if uint64(len(record)) > p.WritableSpace {
		return errors.New("not enough space")
	}

	pointers := p.RecordMap[rowID]
	if len(pointers) == 0 {
		pointers = make([]uint64, 2)
	}

	pointers[0] = uint64(len(p.Data))
	pointers[1] = uint64(len(record))
	p.RecordMap[rowID] = pointers

	p.Data = append(p.Data, record...)
	p.WritableSpace -= uint64(len(record))
	p.FreeSpaceOffset += uint64(len(record))

	return nil
}

func (p *Page) deleteRecord(rowID int64) error {
	pointers, exists := p.RecordMap[rowID]
	if !exists {
		return errors.New("record not found")
	}

	delete(p.RecordMap, rowID)

	p.WritableSpace += pointers[1]
	p.FreeSpaceOffset -= pointers[1]

	return nil
}
