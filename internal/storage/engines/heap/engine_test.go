package heap

import (
	"testing"

	"github.com/platonoff-dev/coredb/internal/storage/pager"
	"github.com/stretchr/testify/require"
)

func prepareEmptyEngine(t *testing.T) *Engine {
	t.Helper()

	pageManager := pager.FilePageManager{
		File:     &pager.MemoryFile{Data: make([]byte, 0)},
		Header:   pager.DBHeader{},
		PageSize: uint32(4096),
	}
	engine := &Engine{
		pageManager: pageManager,
	}

	id1, err := engine.pageManager.Allocate()
	require.NoError(t, err)

	id2, err := engine.pageManager.Allocate()
	require.NoError(t, err)

	require.NoError(t, engine.writePage(&Page{
		ID:              id1,
		Type:            pager.PageTypeHeap,
		FreeSpaceOffset: 0,
		WritableSpace:   4096 - HeaderSize,
		NextPageID:      id2,
		Data:            make([]byte, 0),
		RecordMap:       make(map[int64][]uint64),
	}))

	require.NoError(t, engine.writePage(&Page{
		ID:              id2,
		Type:            pager.PageTypeHeap,
		FreeSpaceOffset: 0,
		WritableSpace:   4096 - HeaderSize,
		Data:            make([]byte, 0),
		RecordMap:       make(map[int64][]uint64),
	}))

	engine.tableMetadata = TableMetadata{
		HeadPageID: id1,
	}

	return engine
}

func prepareEngineWithRecords(t *testing.T) *Engine {
	t.Helper()

	engine := prepareEmptyEngine(t)

	// Insert some records into the engine
	err := engine.Insert(Record{RowID: 1, Data: []byte("Record 1")})
	require.NoError(t, err)

	err = engine.Insert(Record{RowID: 2, Data: []byte("Record 2")})
	require.NoError(t, err)

	err = engine.Insert(Record{RowID: 3, Data: []byte("Record 3")})
	require.NoError(t, err)

	return engine
}

func TestGetRecord(t *testing.T) {
	cases := []struct {
		engine       *Engine
		setupFunc    func(*testing.T, *Engine)
		validateFunc func(*testing.T, *Engine)
		name         string
	}{
		{
			name:   "GetRecord_Basic",
			engine: prepareEngineWithRecords(t),
			setupFunc: func(t *testing.T, e *Engine) {
				t.Helper()
			},

			validateFunc: func(t *testing.T, e *Engine) {
				t.Helper()

				record, err := e.Get(1)
				require.NoError(t, err)
				require.Equal(t, []byte("Record 1"), record.Data)

				record, err = e.Get(2)
				require.NoError(t, err)
				require.Equal(t, []byte("Record 2"), record.Data)

				_, err = e.Get(4) // Non-existent record
				require.Error(t, err)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.setupFunc(t, c.engine)

			c.validateFunc(t, c.engine)
		})
	}
}
