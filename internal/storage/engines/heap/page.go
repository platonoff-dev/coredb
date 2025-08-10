package heap

import (
	"encoding/binary"
	"errors"
)

type Page struct {
	RecordMap  map[int][]int // Row ID to Record Offset mapping and segment length
	Data       []byte
	ID         int
	NextPageID int
	Type       byte

	writableSpacePtr int
	size             int
}

type Record struct {
	Data  []byte
	RowID int
}

func newPage(id int, pageType byte, size int) *Page {
	result := &Page{
		ID:         id,
		Type:       pageType,
		NextPageID: 0,
		RecordMap:  make(map[int][]int),

		size: size,
	}

	result.Data = make([]byte, size-result.requiredHeaderSize())
	result.writableSpacePtr = EndOfPage // End of page

	return result
}

func (p *Page) requiredHeaderSize() int {
	staticVarsSize := 1 + 4*binary.MaxVarintLen64
	recordMapSize := len(p.RecordMap) * 3 * binary.MaxVarintLen64 // 3 uint64s per record

	return staticVarsSize + recordMapSize
}

func (p *Page) writableSpaceLength() int {
	return p.size + p.writableSpacePtr - p.requiredHeaderSize()
}

func (p *Page) MarshalBinary() ([]byte, error) {
	data := make([]byte, p.size)
	offset := 0

	data[offset] = p.Type
	offset++

	n := binary.PutVarint(data[offset:], int64(p.writableSpacePtr))
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

	from := len(data) - len(p.Data)
	copy(data[from:], p.Data)

	return data, nil
}

func (p *Page) UnmarshalBinary(id int, data []byte) error {
	p.ID = id
	offset := 0

	p.Type = data[0]
	offset++

	writableSpacePtr, n := binary.Varint(data[offset:])
	p.writableSpacePtr = int(writableSpacePtr)
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

	p.Data = data[p.requiredHeaderSize():]

	p.size = len(data)

	return nil
}

func (p *Page) listRecords() []Record {
	records := make([]Record, 0, len(p.RecordMap))

	for rowID, pointers := range p.RecordMap {
		if len(pointers) < 2 {
			continue // Invalid record, skip it
		}
		// Pointers are stored as negative offsets from end of data slice (inclusive)
		startPtr := pointers[0]
		endPtr := pointers[1]

		from := len(p.Data) + startPtr + 1
		to := len(p.Data) + endPtr + 1
		if from < 0 || to > len(p.Data) || from > to { // Defensive bounds check
			continue
		}

		record := Record{RowID: rowID, Data: p.Data[from:to]}
		records = append(records, record)
	}

	return records
}

func (p *Page) getRecord(rowID int) (Record, bool) {
	pointers, exists := p.RecordMap[rowID]
	if !exists {
		return Record{}, false
	}

	from := len(p.Data) + pointers[0] + 1
	to := len(p.Data) + pointers[1] + 1
	return Record{
		RowID: rowID,
		Data:  p.Data[from:to],
	}, true
}

// TODO:  When updating a record calculation of free space is little bit different.
// We need to find old record remove it then calculate space for new record and only if it still fits we can write it.
// TODO: Write test first to handle this case.
func (p *Page) setRecord(record Record) error {
	if len(record.Data) > p.writableSpaceLength() {
		return errors.New("not enough space")
	}

	from := p.writableSpacePtr - len(record.Data)
	to := p.writableSpacePtr

	fromIdx := len(p.Data) + from + 1
	toIdx := len(p.Data) + to + 1

	copy(p.Data[fromIdx:toIdx], record.Data)
	p.writableSpacePtr -= len(record.Data)
	p.RecordMap[record.RowID] = []int{from, to}

	return nil
}

func (p *Page) deleteRecord(rowID int) error {
	pointers, exists := p.RecordMap[rowID]
	if !exists {
		return errors.New("record not found")
	}

	delete(p.RecordMap, rowID)
	p.writableSpacePtr += pointers[1] - pointers[0]

	return nil
}
