package cmdutil

import (
	"encoding/json"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// RenderBodyWithInput renders the body (including any templates) for a single request.
//
// The input is the value returned by the iterator which drives the request (e.g. a path parameter).
// When it is a piped line ([]byte), it is made available to the template as input.value,
// otherwise input.value is null (e.g. when a plain value is given via a flag).
//
// Each default is only set when the path does not exist in the rendered body, so values
// set by a template are not overwritten by the defaults.
func RenderBodyWithInput(body *mapbuilder.MapBuilder, input any, defaults ...BodyDefault) (json.RawMessage, error) {
	var pipedInput []byte
	if v, ok := input.([]byte); ok {
		pipedInput = v
	}

	out, err := body.MarshalJSONWithInput(pipedInput)
	if err != nil {
		return nil, err
	}

	for _, d := range defaults {
		if !gjson.GetBytes(out, d.Path).Exists() {
			out, err = sjson.SetBytes(out, d.Path, d.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to set default body value. path=%s, err=%w", d.Path, err)
			}
		}
	}
	return json.RawMessage(out), nil
}

// BodyDefault is a value which is set in the body only if the path does not already exist
type BodyDefault struct {
	Path  string
	Value any
}

// SetBodyValue sets a value in an already rendered body
func SetBodyValue(body json.RawMessage, path string, value any) (json.RawMessage, error) {
	out, err := sjson.SetBytes(body, path, value)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(out), nil
}
