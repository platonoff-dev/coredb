package bitcask

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/platonoff-dev/coredb/kv/engines/common_errors"
)

const (
	roloverSize = 1 * 1024 * 1024 * 1024 // 1 GiB
)

type Engine struct {
	dirPath      string
	keyDir       map[string]KeyPointer
	keyDirLock   sync.RWMutex
	activeFile   *os.File
	rolloverSize int
}

type KeyPointer struct {
	file   string
	offset int
	size   int
}

func New(dirPath string) *Engine {
	return &Engine{
		dirPath:      dirPath,
		keyDir:       make(map[string]KeyPointer),
		rolloverSize: roloverSize,
	}
}

func (e *Engine) Open() error {
	err := os.MkdirAll(e.dirPath, 0750)
	if err != nil {
		return err
	}

	filePath := path.Join(e.dirPath, "active.bitcask.data")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	e.activeFile = file

	// TODO: Looks ugly, refactor
	go func() {
		for {
			time.Sleep(1 * time.Second)
			fileInfo, err := e.activeFile.Stat()
			if err != nil {
				log.Printf("Failed to stat active file: %v", err)
				continue
			}

			size := fileInfo.Size()
			if size >= roloverSize {
				err = e.rolloverActiveFile()
				if err != nil {
					log.Printf("Failed to rollover active file: %v", err)
				}
			}
		}
	}()

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
	dirFiles, err := os.ReadDir(e.dirPath)
	if err != nil {
		return err
	}

	files := make(map[string]bool, len(dirFiles))
	for _, fileInfo := range dirFiles {
		files[fileInfo.Name()] = true
	}

	var filesToMerge []string
	for k := range files {
		if strings.HasSuffix(k, ".hint") {
			continue
		}

		if strings.HasSuffix(k, ".data") {
			hintName := strings.TrimSuffix(k, ".data") + ".hint"
			if _, ok := files[hintName]; !ok {
				filesToMerge = append(filesToMerge, hintName)
			}
		}
	}

	targetFileName := path.Join(e.dirPath, time.Now().Format(time.Now().String()))
	targetFile, err := os.OpenFile(targetFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	for _, fileName := range filesToMerge {
		rawFile, err := os.Open(fileName)
		if err != nil {
			return err
		}

		defer rawFile.Close()

		for {
			entry, err := readEntry(rawFile)
			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}

			if entry.Status != EntryStatusActive {
				continue
			}

			entryData, err := entry.MarshalBinary()
			if err != nil {
				return err
			}
			_, err = targetFile.Write(entryData)
			if err != nil {
				return err
			}

			e.keyDirLock.Lock()
			key := string(entry.Key)
			e.keyDir[key] = KeyPointer{
				file: targetFileName,
				size: e.keyDir[key].size,
			}
			e.keyDirLock.Unlock()
		}
	}

	return nil
}

func (e *Engine) rolloverActiveFile() error {
	e.keyDirLock.Lock()
	defer e.keyDirLock.Unlock()

	fileName := e.activeFile.Name()
	newName := time.Now().String() + ".data"
	e.activeFile.Close()
	err := os.Rename(fileName, newName)
	if err != nil {
		return err
	}

	// TODO: not atomic! Investigate risks
	for _, v := range e.keyDir {
		if v.file == fileName {
			v.file = newName
		}
	}

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}

	e.activeFile = f

	return nil
}

func readEntry(r *os.File) (*Entry, error) {
	var buff = make([]byte, 1+1+4)
	_, err := r.Read(buff)
	if err != nil {
		return nil, err
	}

	status := buff[0]
	keyLength := buff[1]
	valueLength := binary.NativeEndian.Uint32(buff[2:])

	key := make([]byte, keyLength)
	_, err = r.Read(key)
	if err != nil {
		return nil, err
	}

	value := make([]byte, valueLength)
	_, err = r.Read(value)
	if err != nil {
		return nil, err
	}

	entry := &Entry{
		Status: EntryStatus(status),
		Key:    key,
		Value:  value,
	}

	return entry, nil
}
