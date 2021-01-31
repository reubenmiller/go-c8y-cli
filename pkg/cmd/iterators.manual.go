package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

func NewRequestIterator(r c8y.RequestOptions, path iterator.Iterator, body interface{}) *RequestIterator {
	return &RequestIterator{
		Request: r,
		Path:    path,
		Body:    body,
	}
}

func NewRequestFixedPathIterator(r c8y.RequestOptions, path string, body interface{}) *RequestIterator {
	return &RequestIterator{
		Request: r,
		Path:    iterator.NewRepeatIterator(path, 0),
		Body:    body,
	}
}

type RequestIterator struct {
	Request c8y.RequestOptions
	Path    iterator.Iterator
	Body    interface{}
	done    int32
}

func (r *RequestIterator) HasNext() bool {
	return atomic.LoadInt32(&r.done) > 0
}

func (r *RequestIterator) setDone() {
	atomic.AddInt32(&r.done, 1)
}

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
		path, err := r.Path.GetNext()

		if err != nil {
			r.setDone()
			return nil, err
		}
		req.Path = string(path)
	}

	// apply body iterator
	if r.Body != nil && (strings.EqualFold(req.Method, "POST") || strings.EqualFold(req.Method, "PUT")) {
		// iterator body
		bodyContents, err := json.Marshal(r.Body)

		if err != nil {
			r.setDone()
			return nil, err
		}

		// TODO: Find more efficient way rather than converting to and from json
		bodyValue := make(map[string]interface{})

		if err := json.Unmarshal(bodyContents, &bodyValue); err != nil {
			r.setDone()
			return nil, err
		}

		// Apply template
		// if err := setDataTemplateFromFlags(n.cmd, body); err != nil {
		// 	return nil, newUserError("Template error. ", err)
		// }
		// if err := body.Validate(); err != nil {
		// 	return nil, newUserError("Body validation error. ", err)
		// }

		req.Body = bodyValue
	}

	return req, nil
}

func NewBatchFixedPathRequestIterator(cmd *cobra.Command, method string, path string, body interface{}) *RequestIterator {
	// headers
	headers := http.Header{}
	if cmd.Flags().Changed("processingMode") {
		if v, err := cmd.Flags().GetString("processingMode"); err == nil && v != "" {
			headers.Add("X-Cumulocity-Processing-Mode", v)
		}
	}

	req := c8y.RequestOptions{
		Method:       method,
		Header:       headers,
		Body:         body,
		IgnoreAccept: globalFlagIgnoreAccept,
		DryRun:       globalFlagDryRun,
	}

	return NewRequestFixedPathIterator(req, path, body)
}

func NewBatchPathRequestIterator(cmd *cobra.Command, method string, path iterator.Iterator, body interface{}) *RequestIterator {
	// headers
	headers := http.Header{}
	if cmd.Flags().Changed("processingMode") {
		if v, err := cmd.Flags().GetString("processingMode"); err == nil && v != "" {
			headers.Add("X-Cumulocity-Processing-Mode", v)
		}
	}

	req := c8y.RequestOptions{
		Method:       method,
		Header:       headers,
		Body:         body,
		IgnoreAccept: globalFlagIgnoreAccept,
		DryRun:       globalFlagDryRun,
	}

	return NewRequestIterator(req, path, body)
}

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

// NewFlagPipeEnabledStringSlice creates an iterator from a command argument
// or from the pipeline
// It will automatically try to get the value from a String or a StringSlice flag
func NewFlagPipeEnabledStringSlice(cmd *cobra.Command, name string) (iterator.Iterator, error) {
	supportsPipeline := flags.HasValueFromPipeline(cmd, name)
	if cmd.Flags().Changed(name) {

		paths, err := cmd.Flags().GetStringSlice(name)

		if err != nil {
			// fallback to string
			path, err := cmd.Flags().GetString(name)

			if err != nil {
				return nil, err
			}
			paths = append(paths, path)
		}
		if len(paths) > 0 {

			// check if file reference
			if _, err := os.Stat(paths[0]); err == nil {
				iter, err := iterator.NewFileContentsIterator(paths[0])
				if err != nil {
					return nil, err
				}
				return iter, nil
			}

			// return array of results
			return iterator.NewSliceIterator(paths), nil
		}
	} else if supportsPipeline {
		return iterator.NewPipeIterator(func(line []byte) bool {
			return !(bytes.HasPrefix(line, []byte("{")) || bytes.HasPrefix(line, []byte("[")))
		})
	}
	return nil, fmt.Errorf("no input detected")
}
