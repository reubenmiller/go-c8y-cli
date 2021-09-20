package requestiterator

import (
	"bytes"
	"errors"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync/atomic"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

// NewRequestIterator returns an iterator that can be used to send multiple requests until the give iterators in the path/body are exhausted
func NewRequestIterator(customLogger *logger.Logger, r c8y.RequestOptions, path iterator.Iterator, query iterator.Iterator, body interface{}) *RequestIterator {
	if customLogger == nil {
		customLogger = logger.NewDummyLogger("requestiterator")
	}
	reqIter := &RequestIterator{
		Logger:  customLogger,
		Request: r,
		Path:    path,
		Query:   query,
		Body:    body,
	}
	return reqIter
}

// RequestIterator iterates through a c8y rest request with given request options and path iterators
type RequestIterator struct {
	Logger         *logger.Logger
	Request        c8y.RequestOptions
	Path           iterator.Iterator
	Query          iterator.Iterator
	InputIterators flags.RequestInputIterators
	Body           interface{}
	done           int32
}

// HasNext returns true if there the iterator is finished
func (r *RequestIterator) HasNext() bool {
	return atomic.LoadInt32(&r.done) > 0
}

func (r *RequestIterator) setDone() {
	atomic.AddInt32(&r.done, 1)
}

// GetNext return the next request. If error is io.EOF then the iterator is finished
func (r *RequestIterator) GetNext() (*c8y.RequestOptions, interface{}, error) {

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
		DryRun:           r.Request.DryRun,
	}

	var inputLine interface{}
	queryParts := make([]string, 0)

	if v, ok := req.Query.(string); ok {
		if v != "" {
			queryParts = append(queryParts, v)
		}
	}

	// apply path iterator
	if r.Path != nil && !reflect.ValueOf(r.Path).IsNil() {
		path, input, err := r.Path.GetNext()

		if err != nil {
			if !errors.Is(err, cmderrors.ErrNoMatchesFound) {
				r.setDone()
			}
			return nil, nil, err
		}

		if hasUnresolvedVariables(path) {
			return nil, nil, cmderrors.ErrNoMatchesFound
		}

		inputLine = input
		req.Path = string(path)

		if p, err := url.Parse(req.Path); err == nil {
			req.Path = p.Path
			if p.RawQuery != "" {
				queryParts = append(queryParts, p.RawQuery)
			}
		}
	}

	// apply query iterator
	// note: reflection is needed as a simple nil check does not work for interfaces!
	if r.Query != nil && !reflect.ValueOf(r.Query).IsNil() {
		q, input, err := r.Query.GetNext()
		if err != nil {
			if !errors.Is(err, cmderrors.ErrNoMatchesFound) {
				r.setDone()
			}
			return nil, nil, err
		}
		inputLine = input
		// Override any existing query parts to prevent duplication
		queryParts = append([]string{}, strings.ReplaceAll(string(q), " ", "+"))
	}

	if len(queryParts) > 0 {
		req.Query = strings.Join(queryParts, "&")
	}

	r.Logger.Debugf("Input line: %s", inputLine)

	// apply body iterator
	if r.Body != nil && !reflect.ValueOf(r.Body).IsNil() && (strings.EqualFold(req.Method, "POST") || strings.EqualFold(req.Method, "PUT")) {
		// iterator body. Any validation will be run here
		switch v := r.Body.(type) {
		case flags.RawString:
			req.Body = v
		case *os.File:
			req.Body = v
		case *mapbuilder.MapBuilder:
			if v.HasFile() {
				if f, err := v.GetFile(); err != nil {
					return nil, nil, err
				} else {
					req.Body = f
				}
			} else if v.HasRaw() {
				req.Body = v.GetRaw()
			} else {
				bodyContents, err := v.MarshalJSONWithInput(inputLine)
				if err != nil {
					if !errors.Is(err, cmderrors.ErrNoMatchesFound) {
						r.setDone()
					}
					return nil, nil, err
				}

				// TODO: Find more efficient way rather than converting to and from json
				bodyValue := make(map[string]interface{})

				// Note: UnmarshalJSON does not support large numbers by default, so
				// 		 c8y.DecodeJSONBytes should be used instead!
				if err := c8y.DecodeJSONBytes(bodyContents, &bodyValue); err != nil {
					r.setDone()
					return nil, nil, err
				}
				req.Body = bodyValue
			}
		default:
			req.Body = v
		}
	}
	return req, inputLine, nil
}

func hasUnresolvedVariables(v []byte) bool {
	left := bytes.Index(v, []byte("{"))
	right := bytes.Index(v, []byte("}"))
	return left > -1 && right > -1 && right > left
}
