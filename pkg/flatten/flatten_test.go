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

func Test_Flatten(t *testing.T) {
	rawjson := `
	{"name": "testname", "array":[], "fragment": {}}
	`
	// ,
	inputMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(rawjson), &inputMap)
	assert.OK(t, err)
	flatMap, err := Flatten(inputMap, "", DotStyle)
	assert.OK(t, err)
	assert.EqualMarshalJSON(t, flatMap, `{"array":[],"fragment":{},"name":"testname"}`)
}

func Test_FlattenNestedEmptyValues(t *testing.T) {
	rawjson := `
	{"values": { "array":[], "fragment": {} }}
	`
	// ,
	inputMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(rawjson), &inputMap)
	assert.OK(t, err)
	flatMap, err := Flatten(inputMap, "", DotStyle)
	assert.OK(t, err)
	assert.EqualMarshalJSON(t, flatMap, `{"values.array":[],"values.fragment":{}}`)
}
