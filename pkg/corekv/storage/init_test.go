package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	dberrors "github.com/platonoff-dev/coredb/pkg/corekv/errors"
)

func TestInit(t *testing.T) {
	tests := []struct {
		wantErr  error
		setup    func(t *testing.T, filePath string)
		cleanup  func(t *testing.T, filePath string)
		name     string
		filePath string
		pageSize int
	}{
		{
			name:     "new memory file",
			filePath: ":memory",
			pageSize: 4096,
			wantErr:  nil,
		},
		{
			name:     "new file with valid page size",
			filePath: filepath.Join(t.TempDir(), "test.db"),
			pageSize: 4096,
			wantErr:  nil,
		},
		{
			name:     "new file with minimum page size",
			filePath: filepath.Join(t.TempDir(), "test_min.db"),
			pageSize: 512,
			wantErr:  nil,
		},
		{
			name:     "new file with maximum page size",
			filePath: filepath.Join(t.TempDir(), "test_max.db"),
			pageSize: 65536,
			wantErr:  nil,
		},
		{
			name:     "existing valid file",
			filePath: filepath.Join(t.TempDir(), "existing.db"),
			pageSize: 4096,
			wantErr:  nil,
			setup: func(t *testing.T, filePath string) {
				t.Helper()
				// Create a valid database file first
				pm, err := Init(filePath, 4096)
				if err != nil {
					t.Fatalf("Failed to create initial file: %v", err)
				}
				pm.File.Close()
			},
		},
		{
			name:     "existing invalid file",
			filePath: filepath.Join(t.TempDir(), "invalid.db"),
			pageSize: 4096,
			wantErr:  dberrors.ErrInvalidFileFormat,
			setup: func(t *testing.T, filePath string) {
				t.Helper()
				// Create a file with invalid content
				file, err := os.Create(filePath)
				if err != nil {
					t.Fatalf("Failed to create invalid file: %v", err)
				}
				defer file.Close()

				// Write invalid header data
				invalidData := make([]byte, 4096)
				copy(invalidData, "INVALID")
				file.Write(invalidData)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t, tt.filePath)
			}

			if tt.cleanup != nil {
				defer tt.cleanup(t, tt.filePath)
			}

			pm, err := Init(tt.filePath, tt.pageSize)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Init() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				return
			}

			if err != nil {
				t.Errorf("Init() unexpected error = %v", err)
				return
			}

			if pm == nil {
				t.Error("Init() returned nil FilePageManager")
				return
			}

			if !pm.Header.IsValid() {
				t.Error("Header is not valid")
			}

			if string(pm.Header.Magic) != Magic {
				t.Errorf("Header.Magic = %s, want %s", string(pm.Header.Magic), Magic)
			}

			if pm.Header.Version != 1 {
				t.Errorf("Header.Version = %d, want 1", pm.Header.Version)
			}

			if pm.Header.PageSize != tt.pageSize {
				t.Errorf("Header.PageSize = %d, want %d", pm.Header.PageSize, tt.pageSize)
			}

			if pm.Header.FreeListPageID != 1 {
				t.Errorf("Header.FreeListPageID = %d, want 1", pm.Header.FreeListPageID)
			}

			if pm.PageSize != tt.pageSize {
				t.Errorf("PageManager.PageSize = %d, want %d", pm.PageSize, tt.pageSize)
			}

			// Clean up
			if pm.File != nil {
				pm.File.Close()
			}
		})
	}
}

func TestInitValidation(t *testing.T) {
	tests := []struct {
		name        string
		pageSize    int
		expectError bool
	}{
		{
			name:        "page size too small",
			pageSize:    256,
			expectError: true,
		},
		{
			name:        "page size too large",
			pageSize:    131072,
			expectError: true,
		},
		{
			name:        "valid minimum page size",
			pageSize:    512,
			expectError: false,
		},
		{
			name:        "valid maximum page size",
			pageSize:    65536,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm, err := Init(":memory", tt.pageSize)

			//nolint: nestif
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					if pm != nil && pm.File != nil {
						pm.File.Close()
					}
					return
				}
				if !errors.Is(err, dberrors.ErrInvalidFileFormat) {
					t.Errorf("Expected ErrInvalidFileFormat, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if pm == nil {
					t.Error("PageManager is nil")
					return
				}
				pm.File.Close()
			}
		})
	}
}

func TestInitFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	pm, err := Init(dbPath, 4096)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer pm.File.Close()

	// Check file permissions
	fileInfo, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	expectedPerm := os.FileMode(0o600)
	if fileInfo.Mode().Perm() != expectedPerm {
		t.Errorf("File permissions = %v, want %v", fileInfo.Mode().Perm(), expectedPerm)
	}
}

func TestInitCorruptedFile(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "corrupted.db")

	// Create a file with less than minimum header size
	file, err := os.Create(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Write only partial header data (less than required)
	partialData := []byte("CORE")
	_, err = file.Write(partialData)
	if err != nil {
		t.Fatalf("Failed to write partial data: %v", err)
	}
	file.Close()

	_, err = Init(dbPath, 4096)
	if err == nil {
		t.Error("Expected error for corrupted file, got none")
	}
}
