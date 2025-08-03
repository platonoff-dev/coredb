package storage

import (
	"os"

	dberrors "github.com/platonoff-dev/coredb/internal/errors"
	"github.com/platonoff-dev/coredb/internal/storage/pager"
)

const (
	memoryFilePath   = ":memory"
	Magic            = "COREDB"
	dbHeaderPageSize = 4096
)

func Init(
	filePath string,
	pageSize uint32,
) (*pager.FilePageManager, error) {
	file, isNew, err := openFile(filePath)
	if err != nil {
		return nil, err
	}

	var header *pager.DBHeader
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

	return &pager.FilePageManager{
		Header:   *header,
		File:     file,
		PageSize: pageSize,
	}, nil
}

func openFile(filePath string) (file pager.DBFileOperator, isNew bool, err error) {
	if filePath == memoryFilePath {
		isNew = true
		err = nil
		file = &pager.MemoryFile{
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

func initHeader(file pager.DBFileOperator, pageSize uint32) (*pager.DBHeader, error) {
	header := &pager.DBHeader{
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

func readHeader(file pager.DBFileOperator) (*pager.DBHeader, error) {
	headerBuff := make([]byte, dbHeaderPageSize)
	_, err := file.ReadAt(headerBuff, 0)
	if err != nil {
		return nil, err
	}

	header := &pager.DBHeader{}
	err = header.Decode(headerBuff)
	if err != nil {
		return nil, err
	}

	return header, nil
}
