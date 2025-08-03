package pager

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemoryFile_Close(t *testing.T) {
	mf := &MemoryFile{Data: []byte("test data")}
	err := mf.Close()
	if err != nil {
		t.Errorf("Close() should always return nil, got %v", err)
	}
}

func TestMemoryFile_ReadAt(t *testing.T) {
	data := []byte("Hello, World!")
	mf := &MemoryFile{Data: data}

	tests := []struct {
		expectErr  error
		name       string
		expectData string
		offset     int64
		bufSize    int
		expectN    int
	}{
		{
			name:       "read from beginning",
			offset:     0,
			bufSize:    5,
			expectN:    5,
			expectErr:  nil,
			expectData: "Hello",
		},
		{
			name:       "read from middle",
			offset:     7,
			bufSize:    5,
			expectN:    5,
			expectErr:  nil,
			expectData: "World",
		},
		{
			name:       "read entire data",
			offset:     0,
			bufSize:    13,
			expectN:    13,
			expectErr:  nil,
			expectData: "Hello, World!",
		},
		{
			name:       "read beyond end",
			offset:     10,
			bufSize:    10,
			expectN:    3,
			expectErr:  io.EOF,
			expectData: "ld!",
		},
		{
			name:      "negative offset",
			offset:    -1,
			bufSize:   5,
			expectN:   0,
			expectErr: os.ErrInvalid,
		},
		{
			name:      "offset beyond data",
			offset:    20,
			bufSize:   5,
			expectN:   0,
			expectErr: os.ErrInvalid,
		},
		{
			name:      "read at end",
			offset:    13,
			bufSize:   5,
			expectN:   0,
			expectErr: os.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, tt.bufSize)
			n, err := mf.ReadAt(buf, tt.offset)

			if n != tt.expectN {
				t.Errorf("ReadAt() n = %v, want %v", n, tt.expectN)
			}
			if !errors.Is(err, tt.expectErr) {
				t.Errorf("ReadAt() err = %v, want %v", err, tt.expectErr)
			}
			if n > 0 && string(buf[:n]) != tt.expectData {
				t.Errorf("ReadAt() data = %q, want %q", string(buf[:n]), tt.expectData)
			}
		})
	}
}

func TestMemoryFile_WriteAt(t *testing.T) {
	tests := []struct {
		expectErr    error
		name         string
		initialData  []byte
		writeData    []byte
		expectedData []byte
		offset       int64
		expectN      int
	}{
		{
			name:         "write at beginning",
			initialData:  []byte("Hello, World!"),
			writeData:    []byte("Hi"),
			offset:       0,
			expectN:      2,
			expectErr:    nil,
			expectedData: []byte("Hillo, World!"),
		},
		{
			name:         "write in middle",
			initialData:  []byte("Hello, World!"),
			writeData:    []byte("Go"),
			offset:       7,
			expectN:      2,
			expectErr:    nil,
			expectedData: []byte("Hello, Gorld!"),
		},
		{
			name:         "write at end",
			initialData:  []byte("Hello"),
			writeData:    []byte(", World!"),
			offset:       5,
			expectN:      8,
			expectErr:    nil,
			expectedData: []byte("Hello, World!"),
		},
		{
			name:         "write beyond end (expand)",
			initialData:  []byte("Hello"),
			writeData:    []byte("World"),
			offset:       10,
			expectN:      5,
			expectErr:    nil,
			expectedData: []byte("Hello\x00\x00\x00\x00\x00World"),
		},
		{
			name:         "write to empty file",
			initialData:  []byte{},
			writeData:    []byte("Hello"),
			offset:       0,
			expectN:      5,
			expectErr:    nil,
			expectedData: []byte("Hello"),
		},
		{
			name:         "write to empty file with offset",
			initialData:  []byte{},
			writeData:    []byte("Hello"),
			offset:       5,
			expectN:      5,
			expectErr:    nil,
			expectedData: []byte("\x00\x00\x00\x00\x00Hello"),
		},
		{
			name:        "negative offset",
			initialData: []byte("Hello"),
			writeData:   []byte("Hi"),
			offset:      -1,
			expectN:     0,
			expectErr:   os.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf := &MemoryFile{Data: make([]byte, len(tt.initialData))}
			copy(mf.Data, tt.initialData)

			n, err := mf.WriteAt(tt.writeData, tt.offset)

			if n != tt.expectN {
				t.Errorf("WriteAt() n = %v, want %v", n, tt.expectN)
			}
			if !errors.Is(err, tt.expectErr) {
				t.Errorf("WriteAt() err = %v, want %v", err, tt.expectErr)
			}
			if err == nil {
				if len(mf.Data) != len(tt.expectedData) {
					t.Errorf("WriteAt() data length = %v, want %v", len(mf.Data), len(tt.expectedData))
				}
				for i, b := range tt.expectedData {
					if i >= len(mf.Data) || mf.Data[i] != b {
						t.Errorf("WriteAt() data = %v, want %v", mf.Data, tt.expectedData)
						break
					}
				}
			}
		})
	}
}

func TestMemoryFile_Truncate(t *testing.T) {
	tests := []struct {
		expectErr    error
		name         string
		initialData  []byte
		expectedData []byte
		size         int64
	}{
		{
			name:         "truncate to smaller size",
			initialData:  []byte("Hello, World!"),
			size:         5,
			expectErr:    nil,
			expectedData: []byte("Hello"),
		},
		{
			name:         "truncate to larger size",
			initialData:  []byte("Hello"),
			size:         10,
			expectErr:    nil,
			expectedData: []byte("Hello\x00\x00\x00\x00\x00"),
		},
		{
			name:         "truncate to same size",
			initialData:  []byte("Hello"),
			size:         5,
			expectErr:    nil,
			expectedData: []byte("Hello"),
		},
		{
			name:         "truncate to zero",
			initialData:  []byte("Hello, World!"),
			size:         0,
			expectErr:    nil,
			expectedData: []byte{},
		},
		{
			name:         "truncate empty file to larger size",
			initialData:  []byte{},
			size:         5,
			expectErr:    nil,
			expectedData: []byte("\x00\x00\x00\x00\x00"),
		},
		{
			name:        "negative size",
			initialData: []byte("Hello"),
			size:        -1,
			expectErr:   os.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf := &MemoryFile{Data: make([]byte, len(tt.initialData))}
			copy(mf.Data, tt.initialData)

			err := mf.Truncate(tt.size)

			if !errors.Is(err, tt.expectErr) {
				t.Errorf("Truncate() err = %v, want %v", err, tt.expectErr)
			}
			if err == nil {
				if len(mf.Data) != len(tt.expectedData) {
					t.Errorf("Truncate() data length = %v, want %v", len(mf.Data), len(tt.expectedData))
				}
				for i, b := range tt.expectedData {
					if i >= len(mf.Data) || mf.Data[i] != b {
						t.Errorf("Truncate() data = %v, want %v", mf.Data, tt.expectedData)
						break
					}
				}
			}
		})
	}
}

func TestMemoryFile_IntegrationTests(t *testing.T) {
	t.Run("write then read", func(t *testing.T) {
		mf := &MemoryFile{Data: make([]byte, 0)}

		// Write some data
		writeData := []byte("Hello, World!")
		n, err := mf.WriteAt(writeData, 0)
		if err != nil {
			t.Fatalf("WriteAt() failed: %v", err)
		}
		if n != len(writeData) {
			t.Fatalf("WriteAt() n = %v, want %v", n, len(writeData))
		}

		// Read it back
		readBuf := make([]byte, len(writeData))
		n, err = mf.ReadAt(readBuf, 0)
		if err != nil {
			t.Fatalf("ReadAt() failed: %v", err)
		}
		if n != len(writeData) {
			t.Fatalf("ReadAt() n = %v, want %v", n, len(writeData))
		}
		if string(readBuf) != string(writeData) {
			t.Errorf("ReadAt() data = %q, want %q", string(readBuf), string(writeData))
		}
	})

	t.Run("multiple writes and reads", func(t *testing.T) {
		mf := &MemoryFile{Data: make([]byte, 0)}

		// Write at different offsets
		_, err := mf.WriteAt([]byte("Hello"), 0)
		require.NoError(t, err)
		_, err = mf.WriteAt([]byte("World"), 7)
		require.NoError(t, err)
		_, err = mf.WriteAt([]byte(", "), 5)
		require.NoError(t, err)

		// Read the result
		readBuf := make([]byte, 12)
		n, err := mf.ReadAt(readBuf, 0)
		if err != nil {
			t.Fatalf("ReadAt() failed: %v", err)
		}
		expected := "Hello, World"
		if string(readBuf[:n]) != expected {
			t.Errorf("ReadAt() data = %q, want %q", string(readBuf[:n]), expected)
		}
	})

	t.Run("truncate then write and read", func(t *testing.T) {
		mf := &MemoryFile{Data: []byte("Initial data that will be truncated")}

		// Truncate to smaller size
		err := mf.Truncate(7)
		if err != nil {
			t.Fatalf("Truncate() failed: %v", err)
		}

		// Write new data
		_, err = mf.WriteAt([]byte(" test"), 7)
		if err != nil {
			t.Fatalf("WriteAt() failed: %v", err)
		}

		// Read the result
		readBuf := make([]byte, 12)
		n, err := mf.ReadAt(readBuf, 0)
		if err != nil {
			t.Fatalf("ReadAt() failed: %v", err)
		}
		expected := "Initial test"
		if string(readBuf[:n]) != expected {
			t.Errorf("ReadAt() data = %q, want %q", string(readBuf[:n]), expected)
		}
	})
}
