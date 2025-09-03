package heap

import "errors"

type HeapEngine struct {
	pager FilePageManager
}

func NewHeapEngine() *HeapEngine {
	return &HeapEngine{}
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
