package iterator

import "github.com/reubenmiller/go-c8y-cli/pkg/timestamp"

// NewRelativeTimeIterator returns a relative time iterator which can generate timestamps based on time.Now when the value is retrieved
func NewRelativeTimeIterator(relative string, encode bool) *FuncIterator {
	next := func(i int64) (string, error) {
		return timestamp.TryGetTimestamp(relative, encode)
	}
	return NewFuncIterator(next, 0)
}

// NewRelativeDateIterator returns a relative date iterator which can generate dates based on time.Now when the value is retrieved
func NewRelativeDateIterator(relative string, encode bool, layout string) *FuncIterator {
	next := func(i int64) (string, error) {
		return timestamp.TryGetDate(relative, encode, layout)
	}
	return NewFuncIterator(next, 0)
}
