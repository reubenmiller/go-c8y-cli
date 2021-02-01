package iterator

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Iterator is a simple interfact where the next value can be returned.
type Iterator interface {
	GetNext() (line []byte, err error)
}

type UntypedIterator interface {
	GetNext() (out interface{}, err error)
}

func toJSON(i Iterator) (line []byte, err error) {
	v, err := i.GetNext()
	if err != nil {
		if err == io.EOF {
			return []byte("null"), err
		}
		return nil, err
	}

	buff := bytes.NewBufferString("\"")
	buff.Write(bytes.TrimSpace(v))
	buff.WriteByte('"')
	return buff.Bytes(), nil
}

// NewCompositeStringIterator create a new string iterator built from an already existing iterator
func NewCompositeStringIterator(iterator Iterator, format string) *CompositeIterator {
	return &CompositeIterator{
		iterator: iterator,
		format:   format,
	}
}

type CompositeIterator struct {
	format   string
	iterator Iterator
}

// GetNext will count through the values and return them one by one
func (i *CompositeIterator) GetNext() (line []byte, err error) {

	nextValue, err := i.iterator.GetNext()

	if err != nil {
		return line, err
	}

	if strings.Contains(i.format, "%") {
		return []byte(fmt.Sprintf(i.format, nextValue)), nil
	}
	return []byte(i.format), nil

}
