package pager

import (
	"encoding"
)

const (
	PageTypeBTreeInternal = iota + 1
	PageTypeBTreeLeaf
	PageTypeHeap
)

type Page interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}
