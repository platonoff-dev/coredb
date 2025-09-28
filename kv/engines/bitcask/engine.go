package bitcask

import (
	"encoding/binary"
	"errors"
	"os"
	"path"
)

type Engine struct {
	dirPath    string
	keyDir     map[string]KeyPointer
	activeFile *os.File
}

type KeyPointer struct {
	file   []string
	offset int
	size   int
}

func New(dirPath string) *Engine {
	return &Engine{
		dirPath: dirPath,
		keyDir:  make(map[string]KeyPointer),
	}
}

func (e *Engine) Open() error {
	err := os.MkdirAll(e.dirPath, os.ModePerm)
	if err != nil {
		return err
	}

	filePath := path.Join(e.dirPath, "active.bitcask.cdb")
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}

	e.activeFile = file
	return nil
}

func (e *Engine) Close() error {
	err := e.activeFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) Put(key []byte, value []byte) error {
	var record []byte

	if len(key) > 255 {
		return errors.New("key is too long")
	}

	keyLength := byte(len(key))
	record = append(record, keyLength)

	valueLength := len(value)
	record = binary.NativeEndian.AppendUint32(record, uint32(valueLength))

	record = append(record, key...)
	record = append(record, value...)

	return nil
}

func (e *Engine) Get(key []byte) ([]byte, error) {
	return nil, nil
}

func (e *Engine) Delete(key []byte) error {
	return nil
}

func (e *Engine) Sync() error {
	return nil
}

func (e *Engine) Merge() error {
	return nil
}
