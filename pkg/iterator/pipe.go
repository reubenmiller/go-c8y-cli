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
	Properties []string
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
	if i.opts != nil && jsonUtilities.IsJSONObject(line) {
		if len(i.opts.Properties) > 0 {
			// select first property
			for _, prop := range i.opts.Properties {
				if prop != "" {

					if v := gjson.GetBytes(line, prop); v.Exists() {
						return []byte(v.String()), line, nil
					}
				}
			}
			// stop iterator if not found
			err = io.EOF
		}
	}

	return line, line, err
}

// MarshalJSON return the value in a json compatible value
func (i *PipeIterator) MarshalJSON() (line []byte, err error) {
	return MarshalJSON(i)
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
func NewJSONPipeIterator(in io.Reader, pipeOpts *PipeOptions, filter ...Filter) (Iterator, error) {
	var input io.Reader
	switch v := in.(type) {
	case *os.File:
		// check if there is input (otherwise calling .Peek(1) will hang)
		info, err := v.Stat()
		if err != nil {
			return nil, err
		}

		if info.Mode()&os.ModeCharDevice != 0 {
			return nil, ErrNoPipeInput
		}
		input = v
	case io.Reader:
		input = v
	}
	// info, err := os.Stdin.Stat()
	// if err != nil {
	// 	return nil, err
	// }

	// if info.Mode()&os.ModeCharDevice != 0 {
	// 	return nil, ErrNoPipeInput
	// }

	reader := bufio.NewReader(input)
	if _, err := reader.Peek(1); err != nil {
		return nil, ErrNoPipeInput
	}

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
