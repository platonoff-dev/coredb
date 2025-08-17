package heap

import (
	"errors"

	dberrors "github.com/platonoff-dev/coredb/internal/errors"
	"github.com/platonoff-dev/coredb/internal/storage/pager"
)

type Engine struct {
	pageManager pager.FilePageManager
	headPageID  int
}

func (e *Engine) Insert(record Record) error {
	pageID := e.headPageID
	var writablePage, currentPage *Page
	var err error
	for {
		currentPage, err = e.getPage(pageID)
		if err != nil {
			return err
		}

		if currentPage.writableSpaceLength() >= len(record.Data) {
			writablePage = currentPage
			break
		}

		if currentPage.NextPageID == 0 {
			break
		}
	}

	if writablePage == nil {
		writablePage, err = e.appendPage(currentPage)
		if err != nil {
			return err
		}
	}

	err = writablePage.setRecord(record)
	if err != nil {
		return err
	}

	err = e.writePage(writablePage)
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) Get(rowID int) (Record, error) {
	page, err, ok := e.findPage(func(p *Page) bool {
		_, ok := p.RecordMap[rowID]
		return ok
	})

	if err != nil {
		return Record{}, err
	}

	if ok {
		record, _ := page.getRecord(rowID)
		return record, nil
	}

	return Record{}, dberrors.ErrRecordNotFound
}

func (e *Engine) RangeScan() ([]Record, error) {
	records := []Record{}
	pageID := e.headPageID
	for pageID != 0 {
		page, err := e.getPage(pageID)
		if err != nil {
			return nil, err
		}

		// Collect all records from the current page
		records = append(records, page.listRecords()...)

		pageID = page.NextPageID
	}

	return records, nil
}

func (e *Engine) Update(rowID int, record Record) error {
	page, err, ok := e.findPage(func(p *Page) bool {
		_, exists := p.getRecord(rowID)
		return exists
	})
	if err != nil {
		return err
	}
	if !ok {
		return dberrors.ErrRecordNotFound
	}

	err = page.setRecord(Record{RowID: rowID, Data: record.Data})
	if err != nil {
		return err
	}

	err = e.writePage(page)
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) Delete(rowID int) error {
	page, err, ok := e.findPage(func(p *Page) bool {
		_, exists := p.getRecord(rowID)
		return exists
	})

	if err != nil {
		return err
	}

	if !ok {
		return dberrors.ErrRecordNotFound
	}

	err = page.deleteRecord(rowID)
	if err != nil {
		return err
	}

	err = e.writePage(page)
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) getPage(id int) (*Page, error) {
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

func (e *Engine) appendPage(page *Page) (*Page, error) {
	if page == nil {
		return nil, errors.New("page cannot be nil")
	}

	if page.NextPageID != 0 {
		return nil, errors.New("page already has a next page")
	}

	newPageID, err := e.pageManager.Allocate()
	if err != nil {
		return nil, err
	}

	page.NextPageID = newPageID
	err = e.writePage(page)
	if err != nil {
		return nil, err
	}

	nextPage := newPage(newPageID, pager.PageTypeHeap, e.pageManager.PageSize)
	err = e.writePage(nextPage)
	if err != nil {
		return nil, err
	}

	return nextPage, nil
}

func (e *Engine) findPage(cond func(*Page) bool) (*Page, error, bool) {
	pageID := e.headPageID
	for {
		page, err := e.getPage(pageID)
		if err != nil {
			return nil, err, false
		}

		if cond(page) {
			return page, nil, true
		}

		pageID = page.NextPageID
		if pageID == 0 {
			break
		}
	}

	return nil, nil, false //nolint: nilnil
}
