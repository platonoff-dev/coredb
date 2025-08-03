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
	page, err := e.getPage(e.tableMetadata.HeadPageID)
	if err != nil {
		return err
	}

	for page.WritableSpace < uint64(len(record)) || page.NextPageID == 0 {
		page, err = e.getPage(page.NextPageID)
		if err != nil {
			return err
		}
	}

	if page.NextPageID == 0 {
		newPageID, err := e.pageManager.Allocate()
		if err != nil {
			return err
		}

		page.NextPageID = newPageID
		err = e.writePage(page)
		if err != nil {
			return err
		}

		page = &Page{
			ID:              newPageID,
			FreeSpaceOffset: uint64(e.pageManager.PageSize),
			WritableSpace:   uint64(e.pageManager.PageSize) - HeaderSize,
			NextPageID:      0,
		}
	}

	bakwardOffset := page.FreeSpaceOffset - uint64(len(record))
	copy(page.Data[page.FreeSpaceOffset:], record)
	page.FreeSpaceOffset = bakwardOffset
	page.WritableSpace -= uint64(len(record))
	page.RecordMap[rowID] = bakwardOffset

	err = e.writePage(page)
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) Get(rowID int64) (engines.Record, error) {
	return nil, errors.New("not implemented")
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
