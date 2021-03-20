package requestiterator

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

func Test_RequestIteratorWithBodyIterator(t *testing.T) {
	var err error
	pathIter := iterator.NewRepeatIterator("root/subpath", 0)
	valueIter := iterator.NewSliceIterator([]string{"1", "2"})
	body := mapbuilder.NewInitializedMapBuilder()
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
