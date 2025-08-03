package engines

type Record []byte

type Engine interface {
	Insert(rowID int, record Record) error
	Get(rowID int) (Record, error)
	RangeScan() ([]Record, error)
	Update(rowID int, record Record) error
	Delete(rowID int) error
}
