package heap

import (
	"errors"
)

type Pager interface {
	Read(pageID int) ([]byte, error)
	Write(pageID int, data []byte) error
	Allocate(pageID int) ([]byte, error)
}

type HeapEngine struct {
	pager Pager
}

func (e *HeapEngine) Put(key []byte, value []byte) error {

	return errors.New("not implemented")
}

func (e *HeapEngine) Get(key []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (e *HeapEngine) Delete(key []byte) error {
	return errors.New("not implemented")
}
