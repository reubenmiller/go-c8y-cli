package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// NewRelativeTimeIterator returns a relative time iterator which can generate timestamps based on time.Now when the value is retrieved
func NewRelativeTimeIterator(relative string) *iterator.FuncIterator {
	next := func(i int64) (string, error) {
		return tryGetTimestamp(relative)
	}
	return iterator.NewFuncIterator(next, 0)
}

// NewRequestIterator returns an iterator that can be used to send multiple requests until the give iterators in the path/body are exhausted
func NewRequestIterator(r c8y.RequestOptions, path iterator.Iterator, body interface{}) *RequestIterator {
	reqIter := &RequestIterator{
		Request: r,
		Path:    path,
		Body:    body,
	}
	return reqIter
}

// RequestIterator iterates through a c8y rest request with given request options and path iterators
type RequestIterator struct {
	Request      c8y.RequestOptions
	Path         iterator.Iterator
	NameResolver entityFetcher
	Body         interface{}
	done         int32
}

// HasNext returns true if there the iterator is finished
func (r *RequestIterator) HasNext() bool {
	return atomic.LoadInt32(&r.done) > 0
}

func (r *RequestIterator) setDone() {
	atomic.AddInt32(&r.done, 1)
}

// GetNext return the next request. If error is io.EOF then the iterator is finished
func (r *RequestIterator) GetNext() (*c8y.RequestOptions, error) {

	// TODO: is shallow copy ok here?
	req := &c8y.RequestOptions{
		Host:             r.Request.Host,
		Method:           r.Request.Method,
		Path:             r.Request.Path,
		Query:            r.Request.Query,
		Body:             r.Request.Body,
		FormData:         r.Request.FormData,
		ContentType:      r.Request.ContentType,
		Accept:           r.Request.Accept,
		Header:           r.Request.Header,
		ResponseData:     r.Request.ResponseData,
		NoAuthentication: r.Request.NoAuthentication,
		IgnoreAccept:     r.Request.IgnoreAccept,
		DryRun:           globalFlagDryRun,
	}

	// apply path iterator
	if r.Path != nil {
		path, input, err := r.Path.GetNext()

		if err != nil {
			r.setDone()
			return nil, err
		}

		if r.NameResolver != nil {
			if inputB, ok := input.([]byte); ok {

				// skip lookup if path is a fixed string
				if !bytes.Equal(path, inputB) {
					matches, err := lookupIDByName(r.NameResolver, string(inputB))

					if err != nil {
						r.setDone()
						return nil, err
					}
					if len(matches) == 0 {
						r.setDone()
						return nil, fmt.Errorf("no matching results")
					}
					if len(matches) > 0 {
						if f, ok := r.Path.(*iterator.CompositeIterator); ok {
							path, err = f.GetValueByInput([]byte(matches[0].ID))
						} else {
							path = []byte(matches[0].ID)
						}
					}
				}
			}
		}

		req.Path = string(path)
	}

	// apply body iterator
	if r.Body != nil && (strings.EqualFold(req.Method, "POST") || strings.EqualFold(req.Method, "PUT")) {
		// iterator body. Any validation will be run here
		bodyContents, err := json.Marshal(r.Body)

		if err != nil {
			r.setDone()
			return nil, err
		}

		// TODO: Find more efficient way rather than converting to and from json
		bodyValue := make(map[string]interface{})

		// Note: UnmarshalJSON does not support large numbers by default, so
		// 		 c8y.DecodeJSONBytes should be used instead!
		if err := c8y.DecodeJSONBytes(bodyContents, &bodyValue); err != nil {
			r.setDone()
			return nil, err
		}

		req.Body = bodyValue
	}

	return req, nil
}

// NewFlagFileContents returns iterator which will interate over the lines in a file
func NewFlagFileContents(cmd *cobra.Command, name string) (iterator.Iterator, error) {
	supportsPipeline := flags.HasValueFromPipeline(cmd, name)
	if cmd.Flags().Changed(name) {
		if path, err := cmd.Flags().GetString(name); err == nil && path != "" {

			iter, err := iterator.NewFileContentsIterator(path)

			if err != nil {
				return nil, err
			}
			return iter, nil
		}
	} else if supportsPipeline {
		return iterator.NewPipeIterator()
	}
	return nil, fmt.Errorf("no input detected")
}

func NewPipeOption(name string, required bool) *PipeOption {

	return &PipeOption{
		Name:     name,
		Required: required,
	}
}

type PipeOption struct {
	// Name of the flag
	Name string

	// Required marks the pipeline as required, and will return an error if the data is not found in the pipe or from flags
	Required bool

	// Properties slice of json paths to look at mapping a value to the iterator
	Properties []string

	// TODO: Remove this property
	Property string

	// ResolveByNameType type of resolve by name lookup to use
	ResolveByNameType string

	// IteratorType sets whether the iterator is
	IteratorType string
}

type IteraterType string

const IteraterTypeBody = IteraterType("body")
const IteraterTypePath = IteraterType("path")

func NewPathIterator(cmd *cobra.Command, path string, pipeOpt *flags.PipelineOptions) (iterator.Iterator, error) {
	var pathIter iterator.Iterator
	items, err := flags.NewFlagWithPipeIterator(cmd, pipeOpt)
	if err != nil {
		return nil, err
	}

	if items != nil {
		format := path
		pathVarName := pipeOpt.Name

		if len(pipeOpt.Aliases) > 0 {
			pathVarName = pipeOpt.Aliases[0]
		}

		if strings.Count(format, "{") == 1 && strings.Count(path, "}") == 1 {
			// Don't assume the variable name matches the given name.
			// But if there is only one template variable, then it is safe to assume it is the correct one
			format = path[0:strings.Index(path, "{")] + "%s" + path[strings.Index(path, "}")+1:]
		} else {
			// Only substitute an explicitly the pipeline variable name
			format = strings.ReplaceAll(path, "{"+pathVarName+"}", "%s")
		}
		pathIter = iterator.NewCompositeStringIterator(items, format)
	} else {
		pathIter = iterator.NewRepeatIterator(path, 1)
	}
	return pathIter, nil
}
