package bitcask

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"os"
)

type Iterator struct {
	file   *os.File
	offset int64
}

func (i *Iterator) Next() (*Entry, error) {
	header := make([]byte, 4+8+1+binary.MaxVarintLen64+binary.MaxVarintLen64)
	n, err := i.file.ReadAt(header, i.offset)
	if err != nil {
		return nil, err
	}
	if n != len(header) {
		return nil, errors.New("failed to read entry header")
	}

	offset := 0
	crc := binary.LittleEndian.Uint32(header[offset:])
	offset += 4

	ts := binary.LittleEndian.Uint64(header[offset:])
	offset += 8

	status := EntryStatus(header[offset])
	offset++

	keyLength, n := binary.Uvarint(header[offset:])
	offset += n

	valueLength, n := binary.Uvarint(header[offset:])
	offset += n

	data := make([]byte, keyLength+valueLength)
	n, err = i.file.ReadAt(data, i.offset+int64(len(header)))
	if err != nil {
		return nil, err
	}
	if n != len(data) {
		return nil, errors.New("failed to read entry data")
	}

	i.offset += int64(len(header) + len(data))

	if crc != crc32.Checksum(append(header[4:], data...), crcTable) {
		return nil, errors.New("invalid checksum")
	}

	return &Entry{
		CRC:       crc,
		Timestamp: ts,
		Status:    status,
		Key:       data[:keyLength],
		Value:     data[keyLength:],
	}, nil
}
