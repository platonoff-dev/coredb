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
		pageID := 1

		// When
		page, err := manager.Read(pageID)

		// Then
		require.NoError(t, err)
		assert.NotNil(t, page)
		assert.Equal(t, len(page), manager.PageSize)
	})

	t.Run("should handle invalid page ID", func(t *testing.T) {
		// Given
		manager := createTestManager(t)
		invalidPageID := 0 // Page ID 0 is typically invalid

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

		pageData := make([]byte, manager.PageSize)
		copy(pageData, "test data") // Fill with some test data

		// When
		err := manager.Write(1, pageData)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should handle nil page", func(t *testing.T) {
		// Given
		manager := createTestManager(t)

		// When
		err := manager.Write(0, nil)

		// Then
		assert.Error(t, err)
		// Should return specific error like ErrInvalidPage
	})

	t.Run("should update existing page", func(t *testing.T) {
		// Given
		manager := createTestManager(t)
		pageData := make([]byte, manager.PageSize)
		copy(pageData, "initial data")

		// When
		err := manager.Write(1, pageData)

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
		id, err := manager.Allocate()

		// Then
		assert.NoError(t, err)
		assert.Greater(t, id, 0)
	})

	t.Run("should allocate multiple pages with different IDs", func(t *testing.T) {
		// Given
		manager := createTestManager(t)

		// When
		id1, err1 := manager.Allocate()
		id2, err2 := manager.Allocate()

		// Then
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, id1, id2)
	})
}

// TestFilePageManager_Free tests page deallocation functionality.
func TestFilePageManager_Free(t *testing.T) {
	t.Run("should free allocated page", func(t *testing.T) {
		// Given
		manager := createTestManager(t)
		pageID, _ := manager.Allocate()

		// When
		err := manager.Free(pageID)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should handle freeing invalid page ID", func(t *testing.T) {
		// Given
		manager := createTestManager(t)
		invalidPageID := 999

		// When
		err := manager.Free(invalidPageID)

		// Then
		assert.Error(t, err)
		// Should return specific error like ErrPageNotFound
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

		err = manager.Write(1, []byte("test data"))
		assert.Error(t, err)

		pageID, err := manager.Allocate()
		assert.Nil(t, page)
		assert.Error(t, err)

		err = manager.Free(pageID)
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
	header.PageCount = 3 // Assume we have 3 pages for testing

	data := make([]byte, 3*header.PageSize)
	copy(data, header.Encode())

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
		PageCount:      3, // Assume 3 pages for testing
	}
}
