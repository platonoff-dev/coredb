package storage

import (
	"os"

	dberrors "github.com/platonoff-dev/coredb/pkg/corekv/errors"
	pager2 "github.com/platonoff-dev/coredb/pkg/corekv/storage/pager"
)

const (
	memoryFilePath   = ":memory"
	Magic            = "COREDB"
	dbHeaderPageSize = 4096
)

func Init(
	filePath string,
	pageSize int,
) (*pager2.FilePageManager, error) {
	file, isNew, err := openFile(filePath)
	if err != nil {
		return nil, err
	}

	var header *pager2.DBHeader
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

	return &pager2.FilePageManager{
		Header:   *header,
		File:     file,
		PageSize: pageSize,
	}, nil
}

func openFile(filePath string) (file pager2.DBFileOperator, isNew bool, err error) {
	if filePath == memoryFilePath {
		isNew = true
		err = nil
		file = &pager2.MemoryFile{
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

func initHeader(file pager2.DBFileOperator, pageSize int) (*pager2.DBHeader, error) {
	header := &pager2.DBHeader{
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

func readHeader(file pager2.DBFileOperator) (*pager2.DBHeader, error) {
	headerBuff := make([]byte, dbHeaderPageSize)
	_, err := file.ReadAt(headerBuff, 0)
	if err != nil {
		return nil, err
	}

	header := &pager2.DBHeader{}
	err = header.Decode(headerBuff)
	if err != nil {
		return nil, err
	}

	return header, nil
}
