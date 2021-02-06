package iterator

import (
	"fmt"
	"io"
	"sync/atomic"
)

// RangeIterator returns value sequentially from start to end
type RangeIterator struct {
	start   int64
	end     int64
	step    int64
	current int64
}

// NewRangeIterator creates a new range iterator to step through values from a start to end with step
func NewRangeIterator(start, end, step int64) *RangeIterator {
	return &RangeIterator{
		start:   start,
		end:     end,
		step:    step,
		current: start - 1,
	}
}

// GetNext returns the next value from the range
func (i *RangeIterator) GetNext() (line []byte, input interface{}, err error) {
	nextValue := atomic.AddInt64(&i.current, i.step)
	if nextValue > int64(i.end) {
		err = io.EOF
	} else {
		line = []byte(fmt.Sprintf("%d", nextValue))
	}
	return line, nextValue, err
}

// MarshalJSON return the value in a json compatible value
func (i *RangeIterator) MarshalJSON() (line []byte, err error) {
	return MarshalJSON(i)
}
