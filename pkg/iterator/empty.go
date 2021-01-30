package iterator

import "io"

// EmptyIterator is an empty iterator that always returns no value
type EmptyIterator struct{}

// GetNext always return io.EOF
func (i *EmptyIterator) GetNext() (line []byte, err error) {
	return nil, io.EOF
}

// MarshalJSON return the value in a json compatible value
func (i *EmptyIterator) MarshalJSON() (line []byte, err error) {
	return toJSON(i)
}
