package flatten

import (
	"encoding/json"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
)

func Test(t *testing.T) {
	rawjson := `
	{"test.0.value": "one"}
	`
	flatMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(rawjson), &flatMap)
	assert.OK(t, err)
	nestedJSON, err := Unflatten(flatMap)
	assert.OK(t, err)

	nestedMap := make(map[string]interface{})
	err = json.Unmarshal(nestedJSON, &nestedMap)
	assert.OK(t, err)
	assert.EqualMarshalJSON(t, nestedMap, `{"test":[{"value":"one"}]}`)
}
