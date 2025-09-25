package heap

import "errors"

var (
	ErrPageMarshal   = errors.New("marshal page error")
	ErrPageUnmarshal = errors.New("unmarshal page error")
)

type PointersPage struct {
	ID         int
	NextPageID int
	Pointers   map[int]map[string][]int // map[pageID]map[key][offset, size]
}

func (p *PointersPage) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (p *PointersPage) UnmarshalBinary(data []byte) error {
	return nil
}

type DataPage struct {
	ID              int
	FreeSpaceOffset int
	Data            []byte
}

func (p *DataPage) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (p *DataPage) UnmarshalBinary(data []byte) error {
	return nil
}
