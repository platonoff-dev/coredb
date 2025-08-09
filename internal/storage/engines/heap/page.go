package heap

import (
	"encoding/binary"
	"errors"
)

type Page struct {
	RecordMap       map[int][]int // Row ID to Record Offset mapping and segment length
	Data            []byte
	ID              int
	FreeSpaceOffset int
	WritableSpace   int
	NextPageID      int
	Type            byte
}

type Record struct {
	Data  []byte
	RowID int
}

func newPage(id int, pageType byte, pageSize int) *Page {
	result := &Page{
		ID:              id,
		Type:            pageType,
		FreeSpaceOffset: 0,
		NextPageID:      0,

		RecordMap: make(map[int][]int),
		Data:      make([]byte, 0),
	}

	result.WritableSpace = pageSize - result.requiredHeaderSize()

	return result
}

func (p *Page) requiredHeaderSize() int {
	staticVarsSize := 1 + 4*binary.MaxVarintLen64
	recordMapSize := len(p.RecordMap) * 3 * binary.MaxVarintLen64 // 3 uint64s per record

	return staticVarsSize + recordMapSize
}

func (p *Page) MarshalBinary(size int) ([]byte, error) {
	if size < p.requiredHeaderSize() {
		return nil, errors.New("page: invalid size")
	}

	data := make([]byte, size)
	offset := 0

	data[offset] = p.Type
	offset++

	n := binary.PutVarint(data[offset:], int64(p.FreeSpaceOffset))
	offset += n

	n = binary.PutVarint(data[offset:], int64(p.WritableSpace))
	offset += n

	n = binary.PutVarint(data[offset:], int64(p.NextPageID))
	offset += n

	recordCount := len(p.RecordMap)
	n = binary.PutVarint(data[offset:], int64(recordCount))
	offset += n

	for id, pointers := range p.RecordMap {
		n = binary.PutVarint(data[offset:], int64(id))
		offset += n

		n = binary.PutVarint(data[offset:], int64(pointers[0]))
		offset += n

		n = binary.PutVarint(data[offset:], int64(pointers[1]))
		offset += n
	}

	copy(data[offset:], data)

	return data, nil
}

func (p *Page) UnmarshalBinary(id int, data []byte) error {
	p.ID = id
	offset := 0

	p.Type = data[0]
	offset++

	freeSpaceOffset, n := binary.Varint(data[offset:])
	p.FreeSpaceOffset = int(freeSpaceOffset)
	offset += n

	writableSpace, n := binary.Varint(data[offset:])
	p.WritableSpace = int(writableSpace)
	offset += n

	nextPageID, n := binary.Varint(data[offset:])
	offset += n
	p.NextPageID = int(nextPageID)

	recordMapSize, n := binary.Varint(data[offset:])
	offset += n

	p.RecordMap = make(map[int][]int, recordMapSize)
	for i := 0; i < int(recordMapSize); i++ {
		recordID, n := binary.Varint(data[offset:])
		offset += n

		start, n := binary.Varint(data[offset:])
		offset += n

		length, n := binary.Varint(data[offset:])
		offset += n

		p.RecordMap[int(recordID)] = []int{int(start), int(length)}
	}

	p.Data = data[offset:]
	return nil
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

func (p *Page) getRecord(rowID int) (Record, bool) {
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
	if len(record.Data) > p.WritableSpace {
		return errors.New("not enough space")
	}

	pointers := p.RecordMap[record.RowID]
	if len(pointers) == 0 {
		pointers = make([]int, 2)
	}

	pointers[0] = len(p.Data)
	pointers[1] = len(record.Data)
	p.RecordMap[record.RowID] = pointers

	p.Data = append(p.Data, record.Data...)
	p.WritableSpace -= len(record.Data)
	p.FreeSpaceOffset += len(record.Data)

	return nil
}

func (p *Page) deleteRecord(rowID int) error {
	pointers, exists := p.RecordMap[rowID]
	if !exists {
		return errors.New("record not found")
	}

	delete(p.RecordMap, rowID)

	p.WritableSpace += pointers[1]
	p.FreeSpaceOffset -= pointers[1]

	return nil
}
