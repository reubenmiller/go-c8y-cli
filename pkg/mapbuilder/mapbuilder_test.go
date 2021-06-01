package mapbuilder

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

func Test_BodyWithLargeNumbersWithoutTemplates(t *testing.T) {
	body := NewInitializedMapBuilder()

	input := []byte(`{"value":19.1010101E19}`)
	data := map[string]interface{}{}
	err := c8y.DecodeJSONBytes(input, &data)
	assert.OK(t, err)
	body.SetMap(data)
	assert.EqualMarshalJSON(t, body, `{"value":19.1010101E19}`)
}

func Test_BodyWithLargeNumbersWithTemplates(t *testing.T) {
	t.Skip("Large numbers are not supported in templates as jsonnet converts them to integer notation")

	body := NewInitializedMapBuilder()
	input := []byte(`{"value":19.1010101E19}`)
	data := map[string]interface{}{}
	err := c8y.DecodeJSONBytes(input, &data)
	assert.OK(t, err)
	body.SetMap(data)
	body.AppendTemplate("{}")
	assert.EqualMarshalJSON(t, body, `{"value":19.1010101E19}`)
}
