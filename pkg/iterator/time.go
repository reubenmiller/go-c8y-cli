package iterator

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/timestamp"
)

// NewRelativeTimeIterator returns a relative time iterator which can generate timestamps based on time.Now when the value is retrieved
func NewRelativeTimeIterator(relative string, encode bool, utc bool, format ...string) *FuncIterator {
	next := func(i int64) (string, error) {
		value, err := timestamp.TryGetTimestamp(relative, encode, utc)
		if len(format) > 0 {
			if format[0] != "" {
				value = fmt.Sprintf(format[0], value)
			}
		}
		return value, err
	}
	return NewFuncIterator(next, 0)
}

// NewRelativeDateIterator returns a relative date iterator which can generate dates based on time.Now when the value is retrieved
func NewRelativeDateIterator(relative string, encode bool, layout string, format ...string) *FuncIterator {
	next := func(i int64) (string, error) {
		value, err := timestamp.TryGetDate(relative, encode, layout)
		if len(format) > 0 {
			if format[0] != "" {
				value = fmt.Sprintf(format[0], value)
			}
		}
		return value, err
	}
	return NewFuncIterator(next, 0)
}
