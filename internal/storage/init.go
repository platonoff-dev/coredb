package storage

import (
	"os"

	dberrors "github.com/platonoff-dev/coredb/internal/errors"
)

const (
	memoryFilePath   = ":memory"
	Magic            = "COREDB"
	dbHeaderPageSize = 4096
)

func Init(
	filePath string,
	pageSize uint32,
) (*FilePageManager, error) {
	file, isNew, err := openFile(filePath)
	if err != nil {
		return nil, err
	}

	var header *DBHeader
	if isNew {
		header, err = initHeader(file, pageSize)
		if err != nil {
			return nil, err
		}
	} else {
		header, err = readHeader(file)
		if err != nil {
			return nil, err
		}
	}

	if !header.IsValid() {
		return nil, dberrors.ErrInvalidFileFormat
	}

	return &FilePageManager{
		Header:   header,
		File:     file,
		PageSize: pageSize,
	}, nil
}

func openFile(filePath string) (file DBFileOperator, isNew bool, err error) {
	if filePath == memoryFilePath {
		isNew = true
		err = nil
		file = &MemoryFile{
			Data: make([]byte, 0),
		}
		return
	}

	flag := os.O_RDWR
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		flag |= os.O_CREATE
		isNew = true
	}

	file, err = os.OpenFile(filePath, flag, 0o600)
	return
}

func initHeader(file DBFileOperator, pageSize uint32) (*DBHeader, error) {
	header := &DBHeader{
		Magic:          []byte(Magic),
		Version:        1,
		PageSize:       pageSize,
		FreeListPageID: 1,
	}

	err := file.Truncate(dbHeaderPageSize)
	if err != nil {
		return nil, err
	}

	_, err = file.WriteAt(header.Encode(), 0)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func readHeader(file DBFileOperator) (*DBHeader, error) {
	headerBuff := make([]byte, dbHeaderPageSize)
	_, err := file.ReadAt(headerBuff, 0)
	if err != nil {
		return nil, err
	}

	header := &DBHeader{}
	err = header.Decode(headerBuff)
	if err != nil {
		return nil, err
	}

	return header, nil
}
