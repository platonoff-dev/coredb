// Package bitcask key-value store engine implementation.
package bitcask

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/platonoff-dev/coredb/kv/engines/common_errors"
)

const (
	activeFileName = "active.bitcask.data"
	roloverSize    = 1 * 1024 * 1024 * 1024 // 1 GiB
)

type Engine struct {
	activeFile              *os.File
	activeFileSizeThreshold int64
	activeFileSize          int64

	keyDir     map[string]KeyPointer
	keyDirLock sync.RWMutex

	dirPath string
}

type KeyPointer struct {
	file   string
	offset int
	size   int
}

func New(dirPath string) (*Engine, error) {
	engine := &Engine{
		dirPath:                 dirPath,
		keyDir:                  make(map[string]KeyPointer),
		activeFileSizeThreshold: roloverSize,
	}

	err := os.MkdirAll(dirPath, 0750)
	if err != nil {
		return nil, err
	}

	activeFilePath := filepath.Join(dirPath, activeFileName)
	activeFile, err := os.OpenFile(activeFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	engine.activeFile = activeFile

	stat, err := activeFile.Stat()
	if err != nil {
		return nil, err
	}

	engine.activeFileSize = stat.Size()

	return engine, nil
}

func (e *Engine) Close() error {
	err := e.activeFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) Put(key []byte, value []byte) error {
	// It's not really a limitation, but as we keep all keys in ram we need to limit key sizes.
	// 255 bytes is more than enough for everyone
	if len(key) > 255 {
		return errors.New("key is too long")
	}

	entry := &Entry{
		Status: EntryStatusActive,
		Key:    key,
		Value:  value,
	}

	entryBytes, err := entry.MarshalBinary()
	if err != nil {
		return err
	}

	n, err := e.activeFile.Write(entryBytes)
	if err != nil {
		return err
	}

	if n != len(entryBytes) {
		return errors.New("write wrong number of bytes written")
	}

	e.keyDirLock.Lock()
	e.activeFileSize += int64(n)
	e.keyDir[string(key)] = KeyPointer{
		file:   e.activeFile.Name(),
		offset: int(e.activeFileSize),
		size:   len(entryBytes),
	}
	e.keyDirLock.Unlock()

	return nil
}

func (e *Engine) Get(key []byte) ([]byte, error) {
	e.keyDirLock.RLock()
	pointer, ok := e.keyDir[string(key)]
	if !ok {
		return nil, common_errors.ErrKeyNotFound
	}
	e.keyDirLock.RUnlock()

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
	e.keyDirLock.RLock()
	_, ok := e.keyDir[string(key)]
	if !ok {
		return common_errors.ErrKeyNotFound
	}
	e.keyDirLock.RUnlock()

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

	e.activeFileSize += int64(n)

	e.keyDirLock.Lock()
	delete(e.keyDir, string(key))
	e.keyDirLock.Unlock()
	return nil
}

func (e *Engine) Sync() error {
	return e.activeFile.Sync()
}

func (e *Engine) Merge() error {
	return errors.New("not implemented")
}
