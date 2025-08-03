package engines

import (
	"errors"

	"github.com/platonoff-dev/coredb/internal/storage/pager"
)

type Record []byte

type HeapEngine struct {
	pageManager pager.FilePageManager
}

func NewHeapEngine(pageManager pager.FilePageManager) HeapEngine {
	return HeapEngine{
		pageManager: pageManager,
	}
}

func (e *HeapEngine) Insert(rowID int, record Record) error {
	return errors.New("not implemented")
}

func (e *HeapEngine) Get(rowID int) (Record, error) {
	return nil, errors.New("not implemented")
}

func (e *HeapEngine) RangeScan() ([]Record, error) {
	return nil, errors.New("not implemented")
}

func (e *HeapEngine) Update(rowID int, record Record) error {
	return errors.New("not implemented")
}

func (e *HeapEngine) Delete(rowID int) error {
	return errors.New("not implemented")
}
