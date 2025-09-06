package heap

type Page struct {
	Pointers [][]int64 // List of pairs (offset, size). Position corresponds to position of key in keys list
	Data     []byte
}

func (p *Page) Set(k []byte, v []byte) error {
	
	return nil
}

func (p *Page) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (p *Page) UnmarshalBinary(data []byte) error {
	return nil
}
