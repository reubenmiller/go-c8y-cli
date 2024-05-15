package iterator

import (
	"io"
	"sync/atomic"
)

// BoundIterator is generic iterator which executes a function on every iteration
type BoundIterator struct {
	currentIndex int64 // access atomically (must be defined at the top)
	endIndex     int64
	iter         Iterator
}

// GetNext will count through the values and return them one by one
func (i *BoundIterator) GetNext() (line []byte, input interface{}, err error) {
	nextIndex := atomic.AddInt64(&i.currentIndex, 1)
	if i.endIndex != 0 && nextIndex > i.endIndex {
		err = io.EOF
	} else {
		line, input, err = i.iter.GetNext()
	}
	return line, input, err
}

// IsBound return true if the iterator is bound
func (i *BoundIterator) IsBound() bool {
	return i.endIndex > 0
}

// MarshalJSON return the value in a json compatible value
func (i *BoundIterator) MarshalJSON() (line []byte, err error) {
	return MarshalJSON(i)
}

// NewBoundIterator return an iterator which makes an existing iterator bound
func NewBoundIterator(iter Iterator, max int64) *BoundIterator {
	return &BoundIterator{
		iter:     iter,
		endIndex: max,
	}
}
