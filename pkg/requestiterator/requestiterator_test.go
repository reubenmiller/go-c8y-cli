package requestiterator

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/assert"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

func Test_RequestIteratorWithBodyIterator(t *testing.T) {
	var err error
	pathIter := iterator.NewRepeatIterator("root/subpath", 0)
	valueIter := iterator.NewSliceIterator([]string{"1", "2"})
	body := mapbuilder.NewInitializedMapBuilder(true)
	err = body.Set("nested.value", valueIter)
	assert.OK(t, err)
	options := &c8y.RequestOptions{
		Path: "someother/path",
		Body: body,
	}
	requestIter := NewRequestIterator(nil, *options, pathIter, nil, body)

	var req *c8y.RequestOptions

	req, _, err = requestIter.GetNext()
	assert.OK(t, err)
	assert.True(t, req.Path == "root/subpath")
	assert.EqualMarshalJSON(t, req.Body, `{"nested":{"value":"1"}}`)

	req, _, err = requestIter.GetNext()
	assert.OK(t, err)
	assert.True(t, req.Path == "root/subpath")
	assert.EqualMarshalJSON(t, req.Body, `{"nested":{"value":"2"}}`)
}

func Test_RequestIteratorWithEscapedPathVariables(t *testing.T) {
	var err error
	pathIter := iterator.NewRepeatIterator("foo#bar/sub#other", 0)
	valueIter := iterator.NewSliceIterator([]string{"1", "2"})
	body := mapbuilder.NewInitializedMapBuilder(true)
	err = body.Set("nested.value", valueIter)
	assert.OK(t, err)
	options := &c8y.RequestOptions{
		Body: body,
	}
	requestIter := NewRequestIterator(nil, *options, pathIter, nil, body)

	var req *c8y.RequestOptions

	req, _, err = requestIter.GetNext()
	assert.OK(t, err)
	assert.True(t, req.Path == "foo#bar/sub#other")
	assert.EqualMarshalJSON(t, req.Body, `{"nested":{"value":"1"}}`)

	req, _, err = requestIter.GetNext()
	assert.OK(t, err)
	assert.True(t, req.Path == "foo#bar/sub#other")
	assert.EqualMarshalJSON(t, req.Body, `{"nested":{"value":"2"}}`)
}

func Test_RequestIteratorMultipartPreservesQueryParams(t *testing.T) {
	options := c8y.RequestOptions{
		Method: http.MethodPost,
		Path:   "/inventory/binaries",
		Query:  "name=my-file.txt&description=upload",
		FormData: map[string]io.Reader{
			"file": strings.NewReader("demo"),
		},
	}

	pathIter := iterator.NewRepeatIterator(options.Path, 1)
	requestIter := NewRequestIterator(nil, options, pathIter, nil, nil)

	req, _, err := requestIter.GetNext()
	assert.OK(t, err)

	httpReq := httptest.NewRequest(http.MethodPost, "https://example.com/inventory/binaries", nil)
	httpReq, err = req.PrepareRequest(httpReq)
	assert.OK(t, err)

	queryValues, err := url.ParseQuery(httpReq.URL.RawQuery)
	assert.OK(t, err)
	assert.Equal(t, queryValues.Get("name"), "my-file.txt")
	assert.Equal(t, queryValues.Get("description"), "upload")
}
