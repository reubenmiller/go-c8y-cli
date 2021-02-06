package iterator

import (
	"io"
	"sync/atomic"
)

// RepeatIterator is an empty iterator that always returns no value
type RepeatIterator struct {
	value        string
	endIndex     int64
	currentIndex int64
}

// GetNext will count through the values and return them one by one
func (i *RepeatIterator) GetNext() (line []byte, input interface{}, err error) {
	nextIndex := atomic.AddInt64(&i.currentIndex, 1)
	if i.endIndex != 0 && nextIndex > i.endIndex {
		err = io.EOF
	} else {
		line = []byte(i.value)
	}
	return line, line, err
}

// MarshalJSON return the value in a json compatible value
func (i *RepeatIterator) MarshalJSON() (line []byte, err error) {
	return toJSON(i)
}

// NewRepeatIterator creates a repeater which returns the same value n times
// before returns io.EOF. If n is set to 0, then it will repeat forever
func NewRepeatIterator(value string, n int64) *RepeatIterator {
	return &RepeatIterator{
		value:    value,
		endIndex: n,
	}
}
