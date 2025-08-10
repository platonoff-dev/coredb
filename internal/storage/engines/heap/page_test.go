package heap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPage_MarshalBinary(t *testing.T) {
	tests := []struct {
		expectedErr    error
		validateResult func(t *testing.T, result []byte, err error)
		name           string
		page           Page
	}{
		{
			name: "Valid page with sufficient size",
			page: Page{
				Type:             1,
				writableSpacePtr: 20,
				NextPageID:       3,
				size:             200,
				RecordMap: map[int][]int{
					1: {0, 10},
					2: {10, 15},
				},
			},
			expectedErr: nil,
			validateResult: func(t *testing.T, result []byte, err error) {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if len(result) == 0 {
					t.Fatalf("Resulting byte slice is empty")
				}
			},
		},
		{
			name: "Empty page with no records",
			page: Page{
				Type:       2,
				NextPageID: -1,
				RecordMap:  make(map[int][]int),

				writableSpacePtr: 200,
				size:             100,
			},
			expectedErr: nil,
			validateResult: func(t *testing.T, result []byte, err error) {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if len(result) == 0 {
					t.Fatalf("Resulting byte slice is empty")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.page.MarshalBinary()
			if test.validateResult != nil {
				test.validateResult(t, result, err)
			}
		})
	}
}

func TestPageMarshalUnmarshal(t *testing.T) {
	page := newPage(1, 1, 100)

	data, err := page.MarshalBinary()
	require.NoError(t, err)

	newPage := &Page{}
	err = newPage.UnmarshalBinary(1, data)
	require.NoError(t, err)

	assert.Equal(t, page, newPage, "Unmarshaled page does not match original")
}

func TestPage_listRecords(t *testing.T) {
	p := newPage(1, 1, 128)

	// Prepare page with a couple of records
	rec1 := Record{RowID: 1, Data: []byte("hello")}
	rec2 := Record{RowID: 2, Data: []byte("world")}

	require.NoError(t, p.setRecord(rec1))
	require.NoError(t, p.setRecord(rec2))

	records := p.listRecords()
	assert.Len(t, records, 2)

	// Map for easier assertions
	m := map[int][]byte{}
	for _, r := range records {
		m[r.RowID] = r.Data
	}
	assert.Equal(t, []byte("hello"), m[1])
	assert.Equal(t, []byte("world"), m[2])
}

func TestPage_getRecord(t *testing.T) {
	p := newPage(1, 1, 128)

	rec := Record{RowID: 42, Data: []byte("record-data")}
	require.NoError(t, p.setRecord(rec))

	got, ok := p.getRecord(42)
	require.True(t, ok)
	assert.Equal(t, rec.RowID, got.RowID)
	assert.Equal(t, rec.Data, got.Data)

	_, ok = p.getRecord(100) // non-existent
	assert.False(t, ok)
}

func TestPage_setRecord_NotEnoughSpace(t *testing.T) {
	p := newPage(1, 1, 64) // small page

	large := Record{RowID: 1, Data: make([]byte, 10_000)} // obviously too large
	err := p.setRecord(large)
	assert.Error(t, err)
}

func TestPage_deleteRecord_FreesSpace(t *testing.T) {
	p := newPage(1, 1, 256)

	rec1 := Record{RowID: 1, Data: []byte("abcdefgh")} // 8 bytes
	rec2 := Record{RowID: 2, Data: []byte("ijkl")}     // 4 bytes

	require.NoError(t, p.setRecord(rec1))
	ptrAfterRec1 := p.writableSpacePtr
	require.NoError(t, p.setRecord(rec2))
	ptrAfterRec2 := p.writableSpacePtr
	assert.Less(t, ptrAfterRec2, ptrAfterRec1) // moved further left (more negative)

	// Delete second record, expect space pointer to move back by its size
	require.NoError(t, p.deleteRecord(2))
	assert.Equal(t, ptrAfterRec2+len(rec2.Data)-0, p.writableSpacePtr) // freed space

	// Delete first record
	require.NoError(t, p.deleteRecord(1))
}

func TestPage_setRecord_OverwritesExisting(t *testing.T) {
	p := newPage(1, 1, 256)

	rec1 := Record{RowID: 1, Data: []byte("first")}
	require.NoError(t, p.setRecord(rec1))

	// Overwrite with new data (same row id). Current simplistic implementation just appends another copy.
	updated := Record{RowID: 1, Data: []byte("second-version")}
	require.NoError(t, p.setRecord(updated))

	// getRecord returns the last inserted due to map entry being updated
	got, ok := p.getRecord(1)
	require.True(t, ok)
	assert.Equal(t, []byte("second-version"), got.Data)
}
