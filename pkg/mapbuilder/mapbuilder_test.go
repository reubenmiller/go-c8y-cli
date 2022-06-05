package mapbuilder

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

func Test_BodyWithLargeNumbersWithoutTemplates(t *testing.T) {
	body := NewInitializedMapBuilder(true)

	input := []byte(`{"value":19.1010101E19}`)
	data := map[string]interface{}{}
	err := c8y.DecodeJSONBytes(input, &data)
	assert.OK(t, err)
	body.SetMap(data)
	assert.EqualMarshalJSON(t, body, `{"value":19.1010101E19}`)
}

func Test_BodyWithLargeNumbersWithTemplates(t *testing.T) {
	t.Skip("Large numbers are not supported in templates as jsonnet converts them to integer notation")

	body := NewInitializedMapBuilder(true)
	input := []byte(`{"value":19.1010101E19}`)
	data := map[string]interface{}{}
	err := c8y.DecodeJSONBytes(input, &data)
	assert.OK(t, err)
	body.SetMap(data)
	body.AppendTemplate("{}")
	assert.EqualMarshalJSON(t, body, `{"value":19.1010101E19}`)
}

func Test_BodyBuilder_SJON(t *testing.T) {
	body := NewMapBuilder()
	assert.OK(t, body.SetPath("root.value.0.name", "one"))
	assert.EqualJSON(t, body.BodyRaw, `{"root":{"value":[{"name":"one"}]}}`)
}

func Test_MergeMaps(t *testing.T) {
	body := NewMapBuilder()
	assert.OK(t, body.SetPath("root.value.0.name", "one"))
	body.SetOptionalMap(map[string]interface{}{
		"value2": 2,
	})
	out, err := body.MarshalJSON()
	assert.OK(t, err)

	// body should be untouched
	assert.EqualJSON(t, body.BodyRaw, `{"root":{"value":[{"name":"one"}]}}`)

	// body should be merged
	assert.EqualJSON(t, out, `{"root":{"value":[{"name":"one"}]},"value2":2}`)
}

func Test_ExternalInput(t *testing.T) {
	body := NewMapBuilder()

	iter := iterator.NewRepeatIterator("1", 1)

	assert.OK(t, body.Set("root.value.0.name", "one"))
	assert.OK(t, body.Set("input", iter))
	body.externalInput = []byte(`{"name":"peter"}`)
	// body.TemplateIterator
	out, err := body.MarshalJSON()
	assert.OK(t, err)

	// body should be untouched
	assert.EqualJSON(t, body.BodyRaw, `{"root":{"value":[{"name":"one"}]}}`)

	// body should be merged
	assert.EqualJSON(t, out, `{"root":{"value":[{"name":"one"}]},"input":"1"}`)
}
