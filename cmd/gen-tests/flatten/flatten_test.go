package flatten

import (
	"encoding/json"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/assert"
)

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

func Test_FlattenObjectWithLiteralDotInProperty(t *testing.T) {
	rawJSON := `
	{"2021-03-25T17:57:14.973Z": { "max": 10, "min": 1 }}
	`
	inputMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(rawJSON), &inputMap)
	assert.OK(t, err)
	flatMap, err := Flatten(inputMap, "", DotStyle)
	assert.OK(t, err)
	assert.EqualMarshalJSON(t, flatMap, `{"2021-03-25T17:57:14\\.973Z.max":10,"2021-03-25T17:57:14\\.973Z.min":1}`)
}
