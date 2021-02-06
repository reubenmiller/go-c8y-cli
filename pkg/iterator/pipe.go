package iterator

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"sync"

	"errors"

	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/tidwall/gjson"
)

// ErrNoPipeInput is an error when there is no piped input on standard input
var ErrNoPipeInput = errors.New("iterator: no piped input")

// Filter is a funciton applied on every iteration. Returning False will end the iterator
type Filter func([]byte) bool

// PipeOptions additional options on how to interpret the piped data
type PipeOptions struct {
	// Property name if the input data is json
	Property string
}

// PipeIterator is a thread safe iterator to retrieve the input values from piped standard input
type PipeIterator struct {
	mu     sync.Mutex
	filter Filter
	opts   *PipeOptions
	reader *bufio.Reader
}

// GetNext returns the next line from the pipeline
func (i *PipeIterator) GetNext() (line []byte, input interface{}, err error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	line, err = i.reader.ReadBytes('\n')
	line = bytes.TrimSpace(line)

	if err != nil {
		return line, line, err
	}

	if i.filter != nil {
		if !i.filter(line) {
			err = io.EOF
		}
	}

	// check if json, if so pluck the value from it
	if i.opts != nil {
		if i.opts.Property != "" {
			if jsonUtilities.IsJSONObject(line) {
				if v := gjson.GetBytes(line, i.opts.Property); v.Exists() {
					return []byte(v.String()), line, nil
				}
				err = io.EOF
			}
		}
	}

	return line, line, err
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

// NewJSONPipeIterator returns a new pipe iterator
func NewJSONPipeIterator(pipeOpts *PipeOptions, filter ...Filter) (Iterator, error) {
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
		opts:   pipeOpts,
	}, nil
}
