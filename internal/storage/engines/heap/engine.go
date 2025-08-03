package heap

import (
	"errors"

	"github.com/platonoff-dev/coredb/internal/storage/engines"
	"github.com/platonoff-dev/coredb/internal/storage/pager"
)

type TableMetadata struct {
	HeadPageID uint64
	TailPageID uint64
}

type Engine struct {
	pageManager   pager.FilePageManager
	tableMetadata TableMetadata
}

func NewHeapEngine(pageManager pager.FilePageManager) Engine {
	return Engine{
		pageManager: pageManager,
	}
}

func (e *Engine) Insert(rowID int, record engines.Record) error {
	return errors.New("not implemented")
}

func (e *Engine) Get(rowID int) (engines.Record, error) {
	return nil, errors.New("not implemented")
}

func (e *Engine) RangeScan() ([]engines.Record, error) {
	return nil, errors.New("not implemented")
}

func (e *Engine) Update(rowID int, record engines.Record) error {
	return errors.New("not implemented")
}

func (e *Engine) Delete(rowID int) error {
	return errors.New("not implemented")
}
