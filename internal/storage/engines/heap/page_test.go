package heap

import (
	"testing"
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
