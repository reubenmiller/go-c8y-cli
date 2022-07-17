package iterator

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/assert"
)

func Test_MarshalJSON(t *testing.T) {

	iter := NewRangeIterator(1, 3, 1)

	data := make(map[string]interface{})
	data["data"] = iter

	assert.EqualMarshalJSON(t, data, `{"data":"1"}`)
	assert.EqualMarshalJSON(t, data, `{"data":"2"}`)
	assert.EqualMarshalJSON(t, data, `{"data":"3"}`)

	out, err := json.Marshal(data)

	assert.ErrorType(t, err, io.EOF)
	assert.EqualJSON(t, out, "")
}
