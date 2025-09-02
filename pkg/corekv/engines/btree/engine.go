package btree

import "errors"

type BTreeEngine struct {
}

func NewBTreeEngine() *BTreeEngine {
	return &BTreeEngine{}
}

func (e *BTreeEngine) Put(k []byte, v []byte) error {
	return errors.New("not implemented")
}

func (e *BTreeEngine) Get(k []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (e *BTreeEngine) Delete(k []byte) error {
	return errors.New("not implemented")
}

func (e *BTreeEngine) Scan() ([][][]byte, error) {
	return nil, errors.New("not implemented")
}
