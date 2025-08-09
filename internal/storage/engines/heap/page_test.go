package heap

import (
	"errors"
	"testing"
)

func TestPage_MarshalBinary(t *testing.T) {
	tests := []struct {
		expectedErr    error
		validateResult func(t *testing.T, result []byte, err error)
		name           string
		page           Page
		size           int
	}{
		{
			name: "Valid page with sufficient size",
			page: Page{
				Type:            1,
				FreeSpaceOffset: 20,
				WritableSpace:   100,
				NextPageID:      3,
				RecordMap: map[int][]int{
					1: {0, 10},
					2: {10, 15},
				},
			},
			size:        200,
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
			name: "Invalid size",
			page: Page{
				Type:            1,
				FreeSpaceOffset: 20,
				WritableSpace:   100,
				NextPageID:      3,
				RecordMap:       nil,
			},
			size:        5,
			expectedErr: errors.New("page: invalid size"),
			validateResult: func(t *testing.T, result []byte, err error) {
				if err == nil || err.Error() != "page: invalid size" {
					t.Fatalf("Expected error 'page: invalid size', but got: %v", err)
				}
			},
		},
		{
			name: "Empty page with no records",
			page: Page{
				Type:            2,
				FreeSpaceOffset: 0,
				WritableSpace:   200,
				NextPageID:      -1,
				RecordMap:       make(map[int][]int),
			},
			size:        100,
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
			name: "Page with records exceeding size",
			page: Page{
				Type:            3,
				FreeSpaceOffset: 50,
				WritableSpace:   100,
				NextPageID:      4,
				RecordMap: map[int][]int{
					1: {12, 25},
				},
			},
			size:        30,
			expectedErr: errors.New("page: invalid size"),
			validateResult: func(t *testing.T, result []byte, err error) {
				if err == nil || err.Error() != "page: invalid size" {
					t.Fatalf("Expected error 'page: invalid size', but got: %v", err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.page.MarshalBinary(test.size)
			if test.validateResult != nil {
				test.validateResult(t, result, err)
			}
		})
	}
}
