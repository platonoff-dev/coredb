package pager

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFilePageManager_Read tests the basic read functionality.
func TestFilePageManager_Read(t *testing.T) {
	t.Run("should read existing page successfully", func(t *testing.T) {
		// Given
		manager := createTestManager(t)
		pageID, err := manager.Allocate()
		require.NoError(t, err)

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
		invalidPageID := 999

		// When
		page, err := manager.Read(invalidPageID)

		// Then
		assert.Nil(t, page)
		assert.Error(t, err)
		assert.Equal(t, ErrPageNotExist, err)
	})
}

// TestFilePageManager_Write tests the basic write functionality.
func TestFilePageManager_Write(t *testing.T) {
	t.Run("should write new page successfully", func(t *testing.T) {
		// Given
		manager := createTestManager(t)
		pageID, err := manager.Allocate()
		require.NoError(t, err)

		pageData := make([]byte, manager.PageSize)
		copy(pageData, "test data") // Fill with some test data

		err = manager.Write(pageID, pageData)

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
		pageID, err := manager.Allocate()
		require.NoError(t, err)

		pageData := make([]byte, manager.PageSize)
		copy(pageData, "initial data")

		// When
		err = manager.Write(pageID, pageData)

		// Then
		assert.NoError(t, err)
		// Verify the page was updated in storage
	})

	t.Run("should handle invalid page id", func(t *testing.T) {
		// Given
		manager := createTestManager(t)

		// When
		err := manager.Write(999, []byte{})

		// Then
		assert.Error(t, err)
		assert.Equal(t, ErrPageNotExist, err)
	})
}

// TestFilePageManager_Allocate tests page allocation functionality.
func TestFilePageManager_Allocate(t *testing.T) {
	t.Run("should allocate new page with unique ID", func(t *testing.T) {
		// Given
		manager := createTestManager(t)

		// When
		_, err := manager.Allocate()

		// Then
		assert.NoError(t, err)
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

func createTestManager(t *testing.T) *PageManager {
	f, err := os.OpenFile(t.TempDir()+"/file.kv", os.O_CREATE|os.O_RDWR, 0600)
	require.NoError(t, err)

	return &PageManager{
		File:      f,
		PageSize:  4096,
		PageCount: 0,
	}
}
