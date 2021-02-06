package iterator

import (
	"io"
	"sync/atomic"
)

// FuncIterator is generic iterator which executes a function on every iteration
type FuncIterator struct {
	next         func(int64) (string, error)
	endIndex     int64
	currentIndex int64
}

// GetNext will count through the values and return them one by one
func (i *FuncIterator) GetNext() (line []byte, input interface{}, err error) {
	nextIndex := atomic.AddInt64(&i.currentIndex, 1)
	var nextValue string
	if i.endIndex != 0 && nextIndex > i.endIndex {
		err = io.EOF
	} else {
		nextValue, err = i.next(nextIndex)
		line = []byte(nextValue)
	}
	return line, nextValue, err
}

// MarshalJSON return the value in a json compatible value
func (i *FuncIterator) MarshalJSON() (line []byte, err error) {
	return MarshalJSON(i)
}

// NewFuncIterator return an iterator based on the given function
func NewFuncIterator(next func(int64) (string, error), n int64) *FuncIterator {
	return &FuncIterator{
		next:     next,
		endIndex: n,
	}
}
