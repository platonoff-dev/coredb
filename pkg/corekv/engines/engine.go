package engines

type Engine interface {
	Scan() ([][][]byte, error)
	Get(k []byte) ([]byte, error)
	Put(k []byte, v []byte) error
	Delete(k []byte) error
}
