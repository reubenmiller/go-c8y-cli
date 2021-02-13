package iterator

import "github.com/reubenmiller/go-c8y-cli/pkg/timestamp"

// NewRelativeTimeIterator returns a relative time iterator which can generate timestamps based on time.Now when the value is retrieved
func NewRelativeTimeIterator(relative string) *FuncIterator {
	next := func(i int64) (string, error) {
		return timestamp.TryGetTimestamp(relative)
	}
	return NewFuncIterator(next, 0)
}
