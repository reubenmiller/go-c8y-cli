package iterator

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"sync"

	"errors"
)

// ErrNoPipeInput is an error when there is no piped input on standard input
var ErrNoPipeInput = errors.New("iterator: no piped input")

// Filter is a funciton applied on every iteration. Returning False will end the iterator
type Filter func([]byte) bool

// PipeIterator is a thread safe iterator to retrieve the input values from piped standard input
type PipeIterator struct {
	mu     sync.Mutex
	filter Filter
	reader *bufio.Reader
}

// GetNext returns the next line from the pipeline
func (i *PipeIterator) GetNext() (line []byte, err error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	line, err = i.reader.ReadBytes('\n')
	line = bytes.TrimSpace(line)

	if err != nil {
		return line, err
	}

	if i.filter != nil {
		if !i.filter(line) {
			err = io.EOF
		}
	}
	return line, err
}

// MarshalJSON return the value in a json compatible value
func (i *PipeIterator) MarshalJSON() (line []byte, err error) {
	return toJSON(i)
}

// NewPipeIterator returns a new pipe iterator
func NewPipeIterator(filter ...Filter) (Iterator, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	// if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
	if info.Mode()&os.ModeCharDevice != 0 {
		return nil, ErrNoPipeInput
	}

	reader := bufio.NewReader(os.Stdin)

	var pipelineFilter Filter
	if len(filter) > 0 {
		pipelineFilter = filter[0]
	}

	return &PipeIterator{
		reader: reader,
		filter: pipelineFilter,
	}, nil
}
