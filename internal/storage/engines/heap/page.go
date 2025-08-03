package heap

type page struct {
	Data            []byte
	ID              uint64
	FreeSpaceOffset uint32
	Type            byte
}
