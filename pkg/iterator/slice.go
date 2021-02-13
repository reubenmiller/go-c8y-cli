package iterator

import (
	"io"
	"sync/atomic"
)

// SliceIterator is iterates over a given array
type SliceIterator struct {
	values       []string
	currentIndex int64
}

// GetNext will count through the values and return them one by one
func (i *SliceIterator) GetNext() (line []byte, input interface{}, err error) {
	nextIndex := atomic.AddInt64(&i.currentIndex, 1)

	if nextIndex > int64(len(i.values)) {
		err = io.EOF
	} else {
		line = []byte(i.values[nextIndex-1])
	}
	return line, line, err
}

// MarshalJSON return the value in a json compatible value
func (i *SliceIterator) MarshalJSON() (line []byte, err error) {
	return MarshalJSON(i)
}

// IsBound return true if the iterator is bound
func (i *SliceIterator) IsBound() bool {
	return true
}

// NewSliceIterator creates a repeater which returns the slice items
// before returns io.EOF
func NewSliceIterator(values []string) *SliceIterator {
	return &SliceIterator{
		values: values,
	}
}
