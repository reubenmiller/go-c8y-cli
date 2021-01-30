package iterator

import (
	"bufio"
	"bytes"
	"os"
	"sync"

	"errors"
)

// ErrNoPipeInput is an error when there is no piped input on standard input
var ErrNoPipeInput = errors.New("iterator: no piped input")

// PipeIterator is a thread safe iterator to retrieve the input values from piped standard input
type PipeIterator struct {
	mu     sync.Mutex
	reader *bufio.Reader
}

// GetNext returns the next line from the pipeline
func (i *PipeIterator) GetNext() (line []byte, err error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	line, err = i.reader.ReadBytes('\n')
	return bytes.TrimRight(line, "\n"), err
}

// MarshalJSON return the value in a json compatible value
func (i *PipeIterator) MarshalJSON() (line []byte, err error) {
	return toJSON(i)
}

// NewPipeIterator returns a new pipe iterator
func NewPipeIterator() (Iterator, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	// if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
	if info.Mode()&os.ModeCharDevice != 0 {
		return nil, ErrNoPipeInput
	}

	reader := bufio.NewReader(os.Stdin)

	return &PipeIterator{
		reader: reader,
	}, nil
}
