package heap

import (
	"errors"

	dberrors "github.com/platonoff-dev/coredb/internal/errors"
	"github.com/platonoff-dev/coredb/internal/storage/engines"
	"github.com/platonoff-dev/coredb/internal/storage/pager"
)

type TableMetadata struct {
	HeadPageID int64
	TailPageID int64
}

type Engine struct {
	pageManager   pager.FilePageManager
	tableMetadata TableMetadata
}

func NewHeapEngine(pageManager pager.FilePageManager) Engine {
	return Engine{
		pageManager: pageManager,
	}
}

func (e *Engine) Insert(rowID int64, record engines.Record) error {
	pageID := e.tableMetadata.HeadPageID
	var writablePage *Page
	var currentPage *Page
	for {
		currentPage, err := e.getPage(pageID)
		if err != nil {
			return err
		}

		if currentPage.WritableSpace >= uint64(len(record)) {
			writablePage = currentPage
			break
		}

		if currentPage.NextPageID == 0 {
			break
		}
	}

	if writablePage == nil {
		newPageID, err := e.pageManager.Allocate()
		if err != nil {
			return err
		}

		currentPage.NextPageID = newPageID
		err = e.writePage(currentPage)
		if err != nil {
			return err
		}

		writablePage = &Page{
			ID:              newPageID,
			FreeSpaceOffset: uint64(e.pageManager.PageSize),
			WritableSpace:   uint64(e.pageManager.PageSize) - HeaderSize,
			NextPageID:      0,
			Type:            pager.PageTypeHeap,
			RecordMap:       make(map[int64][]uint64),
		}
		err = e.writePage(writablePage)
		if err != nil {
			return err
		}
	}

	writablePage.RecordMap[rowID] = []uint64{writablePage.FreeSpaceOffset, uint64(len(record))}
	writablePage.FreeSpaceOffset -= uint64(len(record))
	writablePage.WritableSpace -= uint64(len(record))
	copy(writablePage.Data[writablePage.FreeSpaceOffset:], record)
	err := e.writePage(writablePage)
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) Get(rowID int64) (engines.Record, error) {
	pageID := e.tableMetadata.HeadPageID
	for {
		page, err := e.getPage(pageID)
		if err != nil {
			return nil, err
		}

		record, ok := e.getRecord(page, rowID)
		if ok {
			return record, nil
		}

		pageID = page.NextPageID
		if pageID == 0 {
			break
		}
	}

	return nil, dberrors.ErrRecordNotFound
}

func (e *Engine) RangeScan() ([]engines.Record, error) {
	return nil, errors.New("not implemented")
}

func (e *Engine) Update(rowID int, record engines.Record) error {
	return errors.New("not implemented")
}

func (e *Engine) Delete(rowID int) error {
	return errors.New("not implemented")
}

func (e *Engine) getPage(id int64) (*Page, error) {
	pageData, err := e.pageManager.Read(id)
	if err != nil {
		return nil, err
	}

	page := &Page{}
	err = page.UnmarshalBinary(id, pageData)
	if err != nil {
		return nil, err
	}

	if page.Type != pager.PageTypeHeap {
		return nil, dberrors.ErrInvalidPageType
	}

	return page, nil
}

func (e *Engine) writePage(page *Page) error {
	if page == nil {
		return errors.New("page cannot be nil")
	}

	binaryPage, err := page.MarshalBinary()
	if err != nil {
		return err
	}

	err = e.pageManager.Write(page.ID, binaryPage)
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) getRecord(page *Page, rowID int64) ([]byte, bool) {
	pointers, exists := page.RecordMap[rowID]
	if !exists {
		return nil, false
	}

	return page.Data[pointers[0] : pointers[0]+pointers[1]], true
}
