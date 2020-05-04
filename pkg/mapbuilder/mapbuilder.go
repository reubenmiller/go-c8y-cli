package mapbuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-jsonnet"
)

const (
	Separator = "."
)

func evaluateJsonnet(base string, snippets ...string) (string, error) {
	// Create a JSonnet VM
	vm := jsonnet.MakeVM()

	var jsonnetImport string

	jsonnetImport += `
local addIfMissing(obj, srcProp, mixin) = if !std.objectHas(obj, srcProp) then mixin else {};
local addIfHas(obj, srcProp, destProp, value) = if std.objectHas(obj, srcProp) then {[destProp]: value} else {};

# Add if property does not exist or is empty.
# If the property is empty then it will be removed!
local addIfEmptyString(obj, srcProp, mixin) = if !std.objectHas(obj, srcProp) || obj[srcProp] == "" then mixin + {[srcProp]:: false} else {};

# Assert: throw error if given property is not present
local isMandatory(o, prop) = {
	assert std.objectHas(o, prop): prop + " is mandatory",
};
`

	base = strings.TrimSpace(base)

	// Add closing ";" if required
	if strings.HasSuffix(base, "}") {
		base = base + ";"
	}

	jsonnetImport += "\nlocal base = " + base

	jsonnetImport += "\nlocal final = "
	if len(snippets) > 0 {
		jsonnetImport += strings.Join(append([]string{"base"}, snippets...), " + ") + ";"
	} else {
		jsonnetImport += "base" + ";"
	}

	jsonnetImport += "\nfinal"

	if strings.ToLower(os.Getenv("C8Y_JSONNET_DEBUG")) == "true" {
		log.Printf("jsonnet snippet: %s\n", jsonnetImport)
	}

	// evaluate the jsonnet
	out, err := vm.EvaluateSnippet("file", jsonnetImport)
	return out, err
}

func NewMapBuilder() *MapBuilder {
	return &MapBuilder{}
}

func NewMapBuilderWithInit(body map[string]interface{}) *MapBuilder {
	return &MapBuilder{
		body: body,
	}
}

// NewMapBuilderFromJsonnetSnippet returns a new mapper builder from a jsonnet snippet
// See https://jsonnet.org/learning/tutorial.html for details on how to create a snippet
func NewMapBuilderFromJsonnetSnippet(snippet string) (*MapBuilder, error) {
	jsonStr, err := evaluateJsonnet(snippet)

	if err != nil {
		return nil, fmt.Errorf("failed to parse snippet. %w", err)
	}

	return NewMapBuilderFromJSON(jsonStr)
}

// NewMapBuilderFromJSON returns a new mapper builder object created from json
func NewMapBuilderFromJSON(data string) (*MapBuilder, error) {
	body := make(map[string]interface{})
	if err := json.Unmarshal([]byte(data), &body); err != nil {
		return nil, err
	}
	return NewMapBuilderWithInit(body), nil
}

// MapBuilder creates body builder
type MapBuilder struct {
	body map[string]interface{}
}

// MergeJsonnet merges the existing body data with a given jsonnet snippet.
// When reverse is false, then the snippet will be applied to the existing data,
// when reverse is true, then the given snippet will be the base, and the existing data will be applied to the new snippet.
func (b *MapBuilder) MergeJsonnet(snippet string, reverse bool) error {
	existingJSON, err := b.MarshalJSON()

	if err != nil {
		return fmt.Errorf("failed to marshal existing map data to json. %w", err)
	}

	var mergedJSON string
	if reverse {
		mergedJSON, err = evaluateJsonnet(snippet, string(existingJSON))
	} else {
		mergedJSON, err = evaluateJsonnet(string(existingJSON), snippet)
	}

	if err != nil {
		return fmt.Errorf("failed to merge json. %w", err)
	}

	body := make(map[string]interface{})
	if err := json.Unmarshal([]byte(mergedJSON), &body); err != nil {
		return fmt.Errorf("failed to decode json. %w", err)
	}
	b.body = body

	return nil
}

// SetMap sets a new map to the body. This will remove any existing values in the body
func (b *MapBuilder) SetMap(body map[string]interface{}) {
	b.body = body
}

// GetMap returns the body as a map[string]interface{}
func (b MapBuilder) GetMap() map[string]interface{} {
	return b.body
}

// GetMap returns the body as a map[string]interface{}
func (b MapBuilder) Get(key string) interface{} {
	return b.body[key]
}

// Get the value as a string
func (b MapBuilder) GetString(key string) (string, bool) {
	val, ok := b.body[key].(string)
	return val, ok
}

// MarshalJSON returns the body as json
func (b MapBuilder) MarshalJSON() ([]byte, error) {
	if b.body == nil {
		return nil, errors.New("body is uninitialized")
	}
	return json.Marshal(b.body)
}

// Set sets a value to a give dot notation path
func (b *MapBuilder) Set(path string, value interface{}) error {
	if b.body == nil {
		b.body = make(map[string]interface{})
	}
	keys := strings.Split(path, Separator)

	currentMap := b.body

	lastIndex := len(keys) - 1

	for i, key := range keys {
		if key != "" {
			if i != lastIndex {
				if _, ok := currentMap[key]; !ok {
					currentMap[key] = make(map[string]interface{})
				}
				currentMap = currentMap[key].(map[string]interface{})
			} else {
				currentMap[key] = value
			}
		}
	}

	return nil
}
