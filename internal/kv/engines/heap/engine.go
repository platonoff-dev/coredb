package heap

import (
	"errors"
)

var (
	ErrPageOperation     = errors.New("page operation error")
	ErrInvalidPageFormat = errors.New("invalid page format")
	ErrPageNotFound      = errors.New("page not found")
)

type Pager interface {
	Read(pageID int) ([]byte, error)
	Write(pageID int, p []byte) error
	Allocate(pageID int) ([]byte, error)
}

type HeapEngine struct {
	pager            Pager
	pointerHeadID    int
	pointerTailID    int
	pointersIterator *LinkedListIterator
}

// Disk
func (e *HeapEngine) Put(key []byte, value []byte) error {
	pointer, err := e.findPointer(key)
	if err != nil {
		return errors.Join(err, ErrPageOperation)
	}

	if pointer != nil {
		// Update existing value
		// Process:
		// How data size changes? If new size <= old size, we can just overwrite the data
		// If new size > old size, we need to find new space for data. Can be on same page or different page
		// Update pointer to new location if needed
		// Write changes to disk (2 pages, data page and pointer page update)
	} else {
		// Pointer not found. Need to insert new value
		// Process:
		// Find page with enough space to insert value
		// If no page found, allocate new page
		// Insert value into page
		// Update pointer page with new key and pointer to value
		// Write changes to disk (2 pages, data page and pointer page update)
	}

	return errors.New("not implemented")
}

func (e *HeapEngine) Get(key []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (e *HeapEngine) Delete(key []byte) error {
	return errors.New("not implemented")
}

// Function iterates through pages to find a page with condition to be true
// It will return nil slice if pointer not found
func (e *HeapEngine) findPointer(key []byte) ([]int, error) {
	// TODO: How to handle empty DB? What is initialization process?
	pointersPageID := e.pointerHeadID

	for {
		pageData, err := e.pager.Read(pointersPageID)
		if err != nil {
			return nil, errors.Join(ErrPageOperation, err)
		}

		page := &PointersPage{}
		err = page.UnmarshalBinary(pageData)
		if err != nil {
			return nil, errors.Join(ErrPageUnmarshal, err)
		}

		for dataPageID, keysPointers := range page.Pointers {
			pointer, ok := keysPointers[string(key)]
			if !ok {
				continue
			}

			return []int{dataPageID, pointer[0], pointer[1]}, nil
		}

		pointersPageID = page.NextPageID

		if pointersPageID == 0 {
			break
		}
	}

	return nil, nil
}

func (e *HeapEngine) findPageWithFreeSpace(requiredSpace int, pageSize int) (int, error) {
	// What to do with empty ?
	for e.pointersIterator.HasNext() {
		page, err := e.pointersIterator.Next()
		if err != nil {
			return 0, err
		}

		if page == nil {
			break
		}

		for dataPageID, keysPositions := range page.Pointers {
			freeSpace := calculateFreeSpaceOnPage(pageSize, keysPositions)
			if freeSpace > requiredSpace {
				return dataPageID, nil
			}
		}
	}

	return 0, nil
}

func calculateFreeSpaceOnPage(pageSize int, keysPositions map[string][]int) int {
	usedSpace := 0
	for _, position := range keysPositions {
		usedSpace += position[1]
	}

	return pageSize - usedSpace
}
