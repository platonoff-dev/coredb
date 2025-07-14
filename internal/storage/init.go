package storage

import (
	"os"
)

const (
	memoryFilePath   = ":memory"
	Magic            = "COREDB"
	dbHeaderPageSize = 4096
)

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

func Init(
	filePath string,
	pageSize uint32,
) (*FilePageManager, error) {
	file, isNew, err := openFile(filePath)
	if err != nil {
		return nil, err
	}

	if isNew {
		// Initialize the file with a header or any necessary initial data
		err := file.Truncate(dbHeaderPageSize)
		if err != nil {
			return nil, err
		}

		header := DBHeader{
			Magic:          []byte(Magic),
			Version:        1,
			PageSize:       pageSize,
			FreeListPageID: 1,
		}
		_, err = file.WriteAt(header.Encode(), 0)
		if err != nil {
			return nil, err
		}
	}

	return &FilePageManager{
		File:     file,
		PageSize: pageSize,
	}, nil
}
