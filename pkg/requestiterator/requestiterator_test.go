package requestiterator

import (
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

	// "#" is escaped (else it starts the url fragment), "/" is kept
	req, _, err = requestIter.GetNext()
	assert.OK(t, err)
	assert.EqualMarshalJSON(t, req.Path, `"foo%23bar/sub%23other"`)
	assert.EqualMarshalJSON(t, req.Body, `{"nested":{"value":"1"}}`)

	req, _, err = requestIter.GetNext()
	assert.OK(t, err)
	assert.EqualMarshalJSON(t, req.Path, `"foo%23bar/sub%23other"`)
	assert.EqualMarshalJSON(t, req.Body, `{"nested":{"value":"2"}}`)
}

func Test_RequestIteratorWithAlreadyEscapedPathVariables(t *testing.T) {
	// A path which is already escaped is not escaped again
	pathIter := iterator.NewRepeatIterator("foo%23bar/sub%23other", 1)
	requestIter := NewRequestIterator(nil, c8y.RequestOptions{}, pathIter, nil, nil)

	req, _, err := requestIter.GetNext()
	assert.OK(t, err)
	assert.EqualMarshalJSON(t, req.Path, `"foo%23bar/sub%23other"`)
}
