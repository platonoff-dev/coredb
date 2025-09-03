package mem

import (
	"github.com/platonoff-dev/coredb/internal/kv/engines/eerrors"
)

type MemEngine struct {
	storage map[string][]byte
}

func NewMemEngine() *MemEngine {
	return &MemEngine{
		storage: make(map[string][]byte),
	}
}

func (e *MemEngine) Get(key []byte) ([]byte, error) {
	v, ok := e.storage[string(key)]
	if !ok {
		return nil, eerrors.ErrKeyNotFound
	}
	return v, nil
}

func (e *MemEngine) Put(key []byte, value []byte) error {
	e.storage[string(key)] = value
	return nil
}

func (e *MemEngine) Delete(key []byte) error {
	delete(e.storage, string(key))
	return nil
}
