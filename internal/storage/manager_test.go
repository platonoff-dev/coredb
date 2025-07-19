package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFilePageManager_Read tests the basic read functionality.
func TestFilePageManager_Read(t *testing.T) {
	t.Run("should read existing page successfully", func(t *testing.T) {
		// Given
		manager := createTestManager(t)
		pageID := uint32(1)

		// When
		page, err := manager.Read(pageID)

		// Then
		require.NoError(t, err)
		assert.NotNil(t, page)
		assert.Equal(t, pageID, page.ID)
	})

	t.Run("should handle invalid page ID", func(t *testing.T) {
		// Given
		manager := createTestManager(t)
		invalidPageID := uint32(0) // Page ID 0 is typically invalid

		// When
		page, err := manager.Read(invalidPageID)

		// Then
		assert.Nil(t, page)
		assert.Error(t, err)
		// Should return specific error like ErrInvalidPageID
	})
}

// TestFilePageManager_Write tests the basic write functionality.
func TestFilePageManager_Write(t *testing.T) {
	t.Run("should write new page successfully", func(t *testing.T) {
		// Given
		manager := createTestManager(t)
		page := &RawPage{
			ID:              1,
			FreeSpaceOffset: 100,
			Type:            PageTypeBTreeLeaf,
			Data:            make([]byte, 4096),
		}

		// When
		err := manager.Write(page)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should handle nil page", func(t *testing.T) {
		// Given
		manager := createTestManager(t)

		// When
		err := manager.Write(nil)

		// Then
		assert.Error(t, err)
		// Should return specific error like ErrInvalidPage
	})

	t.Run("should update existing page", func(t *testing.T) {
		// Given
		manager := createTestManager(t)
		existingPage := &RawPage{
			ID:              1,
			FreeSpaceOffset: 200,
			Type:            PageTypeBTreeInternal,
			Data:            make([]byte, 4096),
		}

		// When
		err := manager.Write(existingPage)

		// Then
		assert.NoError(t, err)
		// Verify the page was updated in storage
	})
}

// TestFilePageManager_Allocate tests page allocation functionality.
func TestFilePageManager_Allocate(t *testing.T) {
	t.Run("should allocate new page with unique ID", func(t *testing.T) {
		// Given
		manager := createTestManager(t)

		// When
		page, err := manager.Allocate()

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, page)
		assert.Greater(t, page.ID, uint32(0))
		assert.Equal(t, manager.PageSize, uint32(len(page.Data))) //nolint: gosec
	})

	t.Run("should allocate multiple pages with different IDs", func(t *testing.T) {
		// Given
		manager := createTestManager(t)

		// When
		page1, err1 := manager.Allocate()
		page2, err2 := manager.Allocate()

		// Then
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotNil(t, page1)
		assert.NotNil(t, page2)
		assert.NotEqual(t, page1.ID, page2.ID)
	})
}

// TestFilePageManager_Free tests page deallocation functionality.
func TestFilePageManager_Free(t *testing.T) {
	t.Run("should free allocated page", func(t *testing.T) {
		// Given
		manager := createTestManager(t)
		page, _ := manager.Allocate()

		// When
		err := manager.Free(page.ID)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should handle freeing invalid page ID", func(t *testing.T) {
		// Given
		manager := createTestManager(t)
		invalidPageID := uint32(999)

		// When
		err := manager.Free(invalidPageID)

		// Then
		assert.Error(t, err)
		// Should return specific error like ErrPageNotFound
	})
}

// TestFilePageManager_Integration tests basic workflow.
func TestFilePageManager_Integration(t *testing.T) {
	t.Run("allocate, write, read, free workflow", func(t *testing.T) {
		// Given
		manager := createTestManager(t)

		// When & Then - Allocate
		page, err := manager.Allocate()
		assert.NoError(t, err)
		assert.NotNil(t, page)

		// Modify page data
		page.Type = PageTypeBTreeLeaf
		copy(page.Data, []byte("test data"))

		// Write page
		err = manager.Write(page)
		assert.NoError(t, err)

		// Read page back
		readPage, err := manager.Read(page.ID)
		assert.NoError(t, err)
		assert.Equal(t, page.ID, readPage.ID)
		assert.Equal(t, page.Type, readPage.Type)
		assert.Equal(t, "test data", string(readPage.Data[:9]))

		// Free page
		err = manager.Free(page.ID)
		assert.NoError(t, err)
	})
}

// TestFilePageManager_EdgeCases tests edge cases and error conditions.
func TestFilePageManager_EdgeCases(t *testing.T) {
	t.Run("should handle manager with nil file", func(t *testing.T) {
		// Given
		manager := &FilePageManager{
			Header:   createTestHeader(),
			File:     nil, // nil file
			PageSize: 4096,
		}

		// When & Then
		page, err := manager.Read(1)
		assert.Nil(t, page)
		assert.Error(t, err)
	})
}

// Helper functions

func createTestManager(t *testing.T) *FilePageManager {
	t.Helper()
	return &FilePageManager{
		Header:   createTestHeader(),
		File:     createTestMemoryFile(),
		PageSize: 4096,
	}
}

func createTestMemoryFile() *MemoryFile {
	header := createTestHeader()
	data := header.Encode()
	data = append(data, make([]byte, 8192)...)
	return &MemoryFile{
		Data: data,
	} // 2 pages worth
}

func createTestHeader() DBHeader {
	return DBHeader{
		Magic:          []byte("COREDB\x00\x00"),
		Version:        1,
		PageSize:       4096,
		FreeListPageID: 1,
	}
}
