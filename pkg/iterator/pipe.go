package iterator

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"errors"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/stream"
	"github.com/tidwall/gjson"
)

// ErrNoPipeInput is an error when there is no piped input on standard input
var ErrNoPipeInput = errors.New("iterator: no piped input")

// ErrEmptyPipeInput pipe input is being used but it is empty
var ErrEmptyPipeInput = errors.New("iterator: empty pipe input")

func IsEmptyPipeInputError(err error) bool {
	return strings.Contains(err.Error(), ErrEmptyPipeInput.Error())
}

// Filter is a funciton applied on every iteration. Returning False will end the iterator
type Filter func([]byte) bool

// Validator is a function applied to every iterator. If it returns an error then the error will be passed on.
// This can be used if you want to apply simple validation logic to each item, i.e. validate that it is an id like value or not.
type Validator func([]byte) error

// Formatter a function to transform input pipeline value and returns the formatted value as output
type Formatter func([]byte) []byte

func DummyFormatter(v []byte) []byte {
	return v
}

func NewStringFormatter(format string) Formatter {
	return func(b []byte) []byte {
		return []byte(fmt.Sprintf(format, b))
	}
}

// PipeOptions additional options on how to interpret the piped data
type PipeOptions struct {
	// Property name if the input data is json
	Properties []string

	// AllowEmpty allow pipeline items without any matching properties.
	AllowEmpty bool

	// Validator to be applied on each item
	Validator Validator

	// Formatter
	Formatter Formatter

	// Format simple custom format string
	Format string
}

// PipeIterator is a thread safe iterator to retrieve the input values from piped standard input
type PipeIterator struct {
	mu     sync.Mutex
	filter Filter
	opts   *PipeOptions
	reader *bufio.Reader
	stream *stream.InputStreamer
}

// IsBound return true if the iterator is bound
func (i *PipeIterator) IsBound() bool {
	return true
}

// GetNext returns the next line from the pipeline
func (i *PipeIterator) GetNext() (line []byte, input interface{}, err error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	line, err = i.stream.Read()
	if len(line) > 0 && errors.Is(err, io.EOF) {
		// Don't return io.EOF if the line includes a value
		// TODO: Ideally this should be changed so the reader of the
		// iterator handles io.EOF correctly (e.g. if there is still a value process it, then
		// react to the io.EOF)
		err = nil
	}

	if i.filter != nil {
		if !i.filter(line) {
			err = io.EOF
		}
	}

	formatter := DummyFormatter

	if i.opts.Format != "" {
		formatter = NewStringFormatter(i.opts.Format)
	}

	if i.opts != nil && i.opts.Formatter != nil {
		formatter = i.opts.Formatter
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
							return formatter([]byte(v.String())), line, nil
						}
						// return raw value (as it might be a number or bool)
						return formatter([]byte(v.Raw)), line, nil
					}
				}
			}
			if i.opts.AllowEmpty && len(bytes.TrimSpace(line)) > 0 {
				// allow empty values, as the user can use the raw piped data in a template
				return []byte(""), line, nil
			}

			// stop iterator if not found (and clear line as some consumers will ignore the EOF if there is some remaining data)
			err = io.EOF
			line = []byte("")
		}
	}

	// validate item if no other errors
	if err == nil && i.opts != nil {
		if i.opts.Validator != nil {
			err = i.opts.Validator(line)
		}
	}

	return formatter(line), line, err
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
	if err := PeekReader(reader); err != nil {
		return nil, err
	}

	var pipelineFilter Filter
	if len(filter) > 0 {
		pipelineFilter = filter[0]
	}

	return &PipeIterator{
		reader: reader,
		stream: &stream.InputStreamer{
			Buffer: reader,
		},
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
	if err := PeekReader(reader); err != nil {
		return nil, err
	}

	var pipelineFilter Filter
	if len(filter) > 0 {
		pipelineFilter = filter[0]
	}

	return &PipeIterator{
		reader: reader,
		stream: &stream.InputStreamer{
			Buffer: reader,
		},
		filter: pipelineFilter,
		opts:   pipeOpts,
	}, nil
}

// PeekReader check if the reader contains empty input or not
// An error will be returned if the reader does not contain any data, or the first character
// is whitespace
func PeekReader(r *bufio.Reader) error {
	peek, err := r.Peek(1)
	if err != nil {
		if err == io.EOF {
			return ErrEmptyPipeInput
		}
		return err
	}
	// check first character contains only whitespace
	if len(bytes.Trim(peek, "\n\r")) == 0 {
		// Treat input starting with an empty line like
		// no pipe input, as when it runs in a cronjob
		// stdin starts with a new lineline char
		return ErrNoPipeInput
	}
	return nil
}
