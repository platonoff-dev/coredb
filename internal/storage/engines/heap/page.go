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

type Record struct {
	Data  []byte
	RowID int64
}

func (p *Page) listRecords() []Record {
	records := make([]Record, 0, len(p.RecordMap))

	for rowID, pointers := range p.RecordMap {
		if len(pointers) < 2 {
			continue // Invalid record, skip it
		}

		start := pointers[0]
		length := pointers[1]

		record := Record{
			RowID: rowID,
			Data:  p.Data[start : start+length],
		}
		records = append(records, record)
	}

	return records
}

func (p *Page) getRecord(rowID int64) (Record, bool) {
	pointers, exists := p.RecordMap[rowID]
	if !exists {
		return Record{}, false
	}

	return Record{
		RowID: rowID,
		Data:  p.Data[pointers[0] : pointers[0]+pointers[1]],
	}, true
}

// TODO:  When updating a record calclulation of frespace is little bit different.
// We need to find old record remove it then calculate space for new record and only if it still fits we can write it.
// TODO: Write test first to handle this case.
func (p *Page) setRecord(record Record) error {
	if uint64(len(record.Data)) > p.WritableSpace {
		return errors.New("not enough space")
	}

	pointers := p.RecordMap[record.RowID]
	if len(pointers) == 0 {
		pointers = make([]uint64, 2)
	}

	pointers[0] = uint64(len(p.Data))
	pointers[1] = uint64(len(record.Data))
	p.RecordMap[record.RowID] = pointers

	p.Data = append(p.Data, record.Data...)
	p.WritableSpace -= uint64(len(record.Data))
	p.FreeSpaceOffset += uint64(len(record.Data))

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
