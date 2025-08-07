package heap

import (
	"errors"

	dberrors "github.com/platonoff-dev/coredb/internal/errors"
	"github.com/platonoff-dev/coredb/internal/storage/pager"
)

type TableMetadata struct {
	HeadPageID int64
}

type Engine struct {
	pageManager   pager.FilePageManager
	tableMetadata TableMetadata
}

func (e *Engine) Insert(record Record) error {
	pageID := e.tableMetadata.HeadPageID
	var writablePage, currentPage *Page
	var err error
	for {
		currentPage, err = e.getPage(pageID)
		if err != nil {
			return err
		}

		if currentPage.WritableSpace >= uint64(len(record.Data)) {
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

func (e *Engine) Get(rowID int64) (Record, error) {
	page, err, ok := e.findPage(func(p *Page) bool {
		_, ok := p.getRecord(rowID)
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
	pageID := e.tableMetadata.HeadPageID
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
		_, exists := p.getRecord(int64(rowID))
		return exists
	})
	if err != nil {
		return err
	}
	if !ok {
		return dberrors.ErrRecordNotFound
	}

	err = page.setRecord(Record{RowID: int64(rowID), Data: record.Data})
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
		_, exists := p.getRecord(int64(rowID))
		return exists
	})

	if err != nil {
		return err
	}

	if !ok {
		return dberrors.ErrRecordNotFound
	}

	err = page.deleteRecord(int64(rowID))
	if err != nil {
		return err
	}

	err = e.writePage(page)
	if err != nil {
		return err
	}

	return nil
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

	binaryPage, err := page.MarshalBinary(int(e.pageManager.PageSize))
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

	newPage := &Page{
		ID:              newPageID,
		FreeSpaceOffset: uint64(e.pageManager.PageSize),
		WritableSpace:   uint64(e.pageManager.PageSize) - HeaderSize,
		NextPageID:      0,
		Type:            pager.PageTypeHeap,
		RecordMap:       make(map[int64][]uint64),
	}
	err = e.writePage(newPage)
	if err != nil {
		return nil, err
	}

	return newPage, nil
}

func (e *Engine) findPage(cond func(*Page) bool) (*Page, error, bool) {
	pageID := e.tableMetadata.HeadPageID
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
