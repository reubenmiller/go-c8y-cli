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

// Validator is a function applied to every iterator. If it returns an error then the error will be passed on.
// This can be used if you want to apply simple validation logic to each item, i.e. validate that it is an id like value or not.
type Validator func([]byte) error

// PipeOptions additional options on how to interpret the piped data
type PipeOptions struct {
	// Property name if the input data is json
	Properties []string

	// AllowEmpty allow pipeline items without any matching properties.
	AllowEmpty bool

	// Validator to be applied on each item
	Validator Validator
}

// PipeIterator is a thread safe iterator to retrieve the input values from piped standard input
type PipeIterator struct {
	mu     sync.Mutex
	filter Filter
	opts   *PipeOptions
	reader *bufio.Reader
}

// IsBound return true if the iterator is bound
func (i *PipeIterator) IsBound() bool {
	return true
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
						// allow for different type
						if v.Type == gjson.String {
							return []byte(v.String()), line, nil
						}
						// return raw value (as it might be a number or bool)
						return []byte(v.Raw), line, nil
					}
				}
			}
			if i.opts.AllowEmpty && len(bytes.TrimSpace(line)) > 0 {
				// allow empty values, as the user can use the raw piped data in a template
				return []byte(""), line, nil
			}

			// stop iterator if not found
			err = io.EOF
		}
	}

	// validate item if no other errors
	if err == nil && i.opts != nil {
		if i.opts.Validator != nil {
			err = i.opts.Validator(line)
		}
	}

	return line, line, err
}

// MarshalJSON return the value in a json compatible value
func (i *PipeIterator) MarshalJSON() (line []byte, err error) {
	return MarshalJSON(i)
}

// NewPipeIterator returns a new pipe iterator
func NewPipeIterator(in io.Reader, filter ...Filter) (Iterator, error) {
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
