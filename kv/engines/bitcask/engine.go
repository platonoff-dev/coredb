package bitcask

import (
	"errors"
	"os"
	"path"

	"github.com/platonoff-dev/coredb/kv/engines/common_errors"
)

type Engine struct {
	dirPath    string
	keyDir     map[string]KeyPointer
	activeFile *os.File
}

type KeyPointer struct {
	file   string
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

	filePath := path.Join(e.dirPath, "active.bitcask.data")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
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
	// It's not really a limitation, but as we keep all keys in ram we need to limit key sizes
	if len(key) > 255 {
		return errors.New("key is too long")
	}

	entry := &Entry{
		Status: EntryStatusActive,
		Key:    key,
		Value:  value,
	}

	data, err := entry.MarshalBinary()
	if err != nil {
		return err
	}

	fileInfo, err := e.activeFile.Stat()
	if err != nil {
		return err
	}
	offset := fileInfo.Size()

	n, err := e.activeFile.Write(data)
	if err != nil {
		return err
	}

	if n != len(data) {
		return errors.New("write wrong number of bytes written")
	}

	e.keyDir[string(key)] = KeyPointer{
		file:   e.activeFile.Name(),
		offset: int(offset),
		size:   len(data),
	}

	return nil
}

func (e *Engine) Get(key []byte) ([]byte, error) {
	pointer, ok := e.keyDir[string(key)]
	if !ok {
		return nil, common_errors.ErrKeyNotFound
	}

	f, err := os.Open(pointer.file)
	if err != nil {
		return nil, err
	}

	var buff = make([]byte, pointer.size)
	n, err := f.ReadAt(buff, int64(pointer.offset))
	if err != nil {
		return nil, err
	}

	if n != pointer.size {
		return nil, errors.New("read wrong number of bytes read")
	}

	entry := &Entry{}
	err = entry.UnmarshalBinary(buff)
	if err != nil {
		return nil, err
	}

	return entry.Value, nil
}

func (e *Engine) Delete(key []byte) error {
	_, ok := e.keyDir[string(key)]
	if !ok {
		return common_errors.ErrKeyNotFound
	}

	deleteEntry := &Entry{
		Status: EntryStatusDeleted,
		Key:    key,
		Value:  nil,
	}
	entryData, err := deleteEntry.MarshalBinary()
	if err != nil {
		return err
	}

	n, err := e.activeFile.Write(entryData)
	if err != nil {
		return err
	}

	if n != len(entryData) {
		return errors.New("write wrong number of bytes written")
	}

	delete(e.keyDir, string(key))
	return nil
}

func (e *Engine) Sync() error {
	return e.activeFile.Sync()
}

func (e *Engine) Merge() error {
	return nil
}
