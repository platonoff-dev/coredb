package bitcask

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
)

type EntryStatus byte

const (
	EntryStatusActive EntryStatus = iota
	EntryStatusDeleted
)

var crcTable = crc32.MakeTable(crc32.Castagnoli)

type Entry struct {
	CRC       uint32
	Timestamp uint64
	Status    EntryStatus

	Key   []byte
	Value []byte
}

func (e *Entry) MarshalBinary() ([]byte, error) {
	if len(e.Key) > 255 {
		return nil, errors.New("key too long")
	}

	// Init
	buffLen := 4 + 8 + 1 + binary.MaxVarintLen64 + binary.MaxVarintLen64 + len(e.Key) + len(e.Value)
	offset := 0
	buff := make([]byte, buffLen)

	buff[0] = byte(e.Status)
	offset++

	binary.LittleEndian.PutUint64(buff[offset:], e.Timestamp)
	offset += 8

	offset += binary.PutUvarint(buff[offset:], uint64(len(e.Key)))

	offset += binary.PutUvarint(buff[offset:], uint64(len(e.Value)))

	copy(buff[offset:], e.Key)
	offset += len(e.Key)

	copy(buff[offset:], e.Value)
	offset += len(e.Value)

	result := buff[:offset]
	checksum := crc32.Checksum(result, crcTable)
	checksumBuff := make([]byte, 4)
	binary.LittleEndian.PutUint32(checksumBuff, checksum)

	return bytes.Join([][]byte{checksumBuff, result}, []byte{}), nil
}

func (e *Entry) UnmarshalBinary(data []byte) error {
	offset := 0

	e.CRC = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	if e.CRC != crc32.Checksum(data[offset:], crcTable) {
		return errors.New("invalid checksum")
	}

	e.Timestamp = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	e.Status = EntryStatus(data[offset])
	offset++

	keyLength, n := binary.Uvarint(data[offset:])
	offset += n

	valueLength, n := binary.Uvarint(data[offset:])
	offset += n

	e.Key = make([]byte, keyLength)
	copy(e.Key, data[offset:offset+int(keyLength)]) //nolint:gosec
	offset += int(keyLength)                        //nolint:gosec

	e.Value = make([]byte, valueLength)
	copy(e.Value, data[offset:offset+int(valueLength)]) //nolint:gosec

	return nil
}
