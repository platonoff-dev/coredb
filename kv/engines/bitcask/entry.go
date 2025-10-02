package bitcask

import (
	"encoding/binary"
	"errors"
)

type EntryStatus byte

const (
	EntryStatusActive EntryStatus = iota
	EntryStatusDeleted
)

type Entry struct {
	CRC       uint32
	Timestamp int64
	Status    EntryStatus

	Key   []byte
	Value []byte
}

func (e *Entry) MarshalBinary() ([]byte, error) {
	if len(e.Key) > 255 {
		return nil, errors.New("key too long")
	}

	buffLen := 1 + 1 + 4 + len(e.Key) + len(e.Value)
	buff := make([]byte, buffLen)

	buff[0] = byte(e.Status)
	buff[1] = byte(len(e.Key))
	binary.NativeEndian.PutUint32(buff[2:6], uint32(len(e.Value)))
	copy(buff[6:6+len(e.Key)], e.Key)
	copy(buff[6+len(e.Key):], e.Value)

	return buff, nil
}

func (e *Entry) UnmarshalBinary(data []byte) error {
	e.Status = EntryStatus(data[0])
	keyLength := data[1]
	valueLength := binary.NativeEndian.Uint32(data[2:6])
	key := make([]byte, keyLength)
	value := make([]byte, valueLength)
	copy(key, data[6:6+int(keyLength)])
	copy(value, data[6+int(keyLength):])

	e.Key = key
	e.Value = value
	return nil
}
