package heap

import (
	"errors"

	"github.com/platonoff-dev/coredb/internal/storage/pager"
)

type Engine struct {
	pageManager pager.FilePageManager
	headPageID  int
}

func (e *Engine) Insert(key []byte, value []byte) error {
	return errors.New("not implemented")
}

func (e *Engine) Get(key []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (e *Engine) RangeScan() ([][][]byte, error) {
	return nil, errors.New("not implemented")
}

func (e *Engine) Update(key int, value []byte) error {
	return nil
}

func (e *Engine) Delete(key []byte) error {
	return errors.New("not implemented")
}
