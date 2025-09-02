package memmap

type MemMapEngine struct {
	storage map[string][]byte
}

func NewMemMapEngine() *MemMapEngine {
	return &MemMapEngine{
		storage: make(map[string][]byte),
	}
}

func (e *MemMapEngine) Get(key []byte) ([]byte, error) {
	return e.storage[string(key)], nil
}

func (e *MemMapEngine) Put(key []byte, value []byte) error {
	e.storage[string(key)] = value
	return nil
}

func (e *MemMapEngine) Delete(key []byte) error {
	delete(e.storage, string(key))
	return nil
}

func (e *MemMapEngine) Scan() ([][][]byte, error) {
	result := make([][][]byte, 0, len(e.storage))
	for k, v := range e.storage {
		pair := [][]byte{[]byte(k), v}
		result = append(result, pair)
	}

	return result, nil
}
