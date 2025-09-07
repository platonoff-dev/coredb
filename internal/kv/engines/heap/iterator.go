package heap

import "errors"

var (
	ErrEndOfIteration = errors.New("end of iteration")
)

type LinkedListIterator struct {
	current *PointersPage
	pager   Pager
}

func (it *LinkedListIterator) HasNext() bool {
	return it.current != nil
}

func (it *LinkedListIterator) Next() (*PointersPage, error) {
	if it.current == nil {
		return nil, nil
	}

	if it.current.NextPageID == 0 {
		it.current = nil
		return it.current, nil
	}

	pageID := it.current.NextPageID
	pageData, err := it.pager.Read(pageID)
	if err != nil {
		return nil, errors.Join(ErrPageOperation, err)
	}

	page := &PointersPage{}
	err = page.UnmarshalBinary(pageData)
	if err != nil {
		return nil, err
	}

	it.current = page
	return it.current, nil
}

func (it *LinkedListIterator) Current() (*PointersPage, error) {
	return it.current, nil
}
