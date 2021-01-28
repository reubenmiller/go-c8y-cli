package mapbuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/sethvargo/go-password/password"
)

const (
	Separator = "."
)

func evaluateJsonnet(base string, imports string, snippets ...string) (string, error) {
	// Create a JSonnet VM
	vm := jsonnet.MakeVM()

	jsonnetImport := imports

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

	debugJsonnet := strings.EqualFold(os.Getenv("C8Y_JSONNET_DEBUG"), "true")

	if debugJsonnet {
		log.Printf("jsonnet snippet: %s\n", jsonnetImport)
	}

	// evaluate the jsonnet
	out, err := vm.EvaluateSnippet("file", jsonnetImport)

	if err != nil {

		if debugJsonnet {
			// Include full template (with injected variables/functions) otherwise the error
			// will report line numbers that the user does not know about
			log.Printf("jsonnet error: %s\n", err)
		}

		err = fmt.Errorf("Could not create json from template. Error: %s", err)
	}
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
	jsonStr, err := evaluateJsonnet(snippet, "")

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
	body          map[string]interface{}
	file          string
	TemplateIndex string

	templateVariables map[string]interface{}
	requiredKeys      []string
	template          string
	validateTemplate  string
}

func (b *MapBuilder) SetTemplate(template string) *MapBuilder {
	b.template = template
	return b
}

func (b *MapBuilder) ApplyTemplate(reverse bool) error {
	return b.MergeJsonnet(b.template, reverse)
}

// MergeJsonnet merges the existing body data with a given jsonnet snippet.
// When reverse is false, then the snippet will be applied to the existing data,
// when reverse is true, then the given snippet will be the base, and the existing data will be applied to the new snippet.
func (b *MapBuilder) MergeJsonnet(snippet string, reverse bool) error {
	var err error
	existingJSON := []byte("{}")

	if b.body != nil {
		existingJSON, err = b.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal existing map data to json. %w", err)
		}
	}

	imports, err := b.GetTemplateVariablesJsonnet()
	if err != nil {
		return err
	}

	var mergedJSON string
	if reverse {
		mergedJSON, err = evaluateJsonnet(snippet, imports, string(existingJSON))
	} else {
		mergedJSON, err = evaluateJsonnet(string(existingJSON), imports, snippet)
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

func (b *MapBuilder) SetTemplateVariables(variables map[string]interface{}) {
	b.templateVariables = variables
}

func generatePassword() string {
	passwordGen, err := password.NewGenerator(&password.GeneratorInput{
		Symbols: "!@#%^()[]*+-_;,.",
	})

	if err != nil {
		return ""
	}

	if res, err := passwordGen.Generate(32, 2, 2, false, false); err == nil {
		return res
	}

	return ""
}

func (b *MapBuilder) GetTemplateVariablesJsonnet() (string, error) {
	if b.templateVariables == nil {
		// template variables have not been defined (not an error)
		return "", nil
	}

	jsonStr, err := json.Marshal(b.templateVariables)

	if err != nil {
		return "", err
	}

	varsHelper := `local var(prop, defaultValue="") = if std.objectHas(vars, prop) then vars[prop] else defaultValue;`

	// Seed random otherwise it will not change with execution
	rand.Seed(time.Now().UTC().UnixNano())

	index := "1"
	if b.TemplateIndex != "" {
		index = b.TemplateIndex
	}

	randomHelper := fmt.Sprintf(`local rand = { index: %s, bool: %t, int: %d, int2: %d, float: %f, float2: %f, float3: %f, float4: %f, password: "%s" };`,
		index,
		rand.Float32() > 0.5,
		rand.Intn(100),
		rand.Intn(100),
		rand.Float32(),
		rand.Float32(),
		rand.Float32(),
		rand.Float32(),
		generatePassword(),
	)
	timeHelper := fmt.Sprintf(`local time = {now: "%s", nowNano: "%s"};`,
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339Nano),
	)
	return fmt.Sprintf("\nlocal vars = %s;\n%s\n%s\n%s\n", jsonStr, varsHelper, randomHelper, timeHelper), nil
}

// SetMap sets a new map to the body. This will remove any existing values in the body
func (b *MapBuilder) SetMap(body map[string]interface{}) {
	b.body = body
}

// SetFile sets the body to the contents of the file path
func (b *MapBuilder) SetFile(path string) {
	b.file = path
}

// GetMap returns the body as a map[string]interface{}
func (b MapBuilder) GetMap() map[string]interface{} {
	return b.body
}

func (b MapBuilder) GetFileContents() *os.File {
	file, err := os.Open(b.file)
	if err != nil {
		log.Printf("failed to open file. %s", err)
		return nil
	}
	return file
}

func (b MapBuilder) GetBody() (interface{}, error) {
	if b.file != "" {
		return os.Open(b.file)
	}
	return b.GetMap(), nil
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

func (b *MapBuilder) SetRequiredKeys(keys ...string) {
	b.requiredKeys = keys
}

// SetValidationTemplate sets the validation template to be applied. The template accepts a jsonnet string
func (b *MapBuilder) SetValidationTemplate(template string) {
	b.validateTemplate = template
}

func (b MapBuilder) validateRequiredKeys() error {
	missingKeys := make([]string, 0)
	for _, key := range b.requiredKeys {
		if _, ok := b.body[key]; !ok {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf("Missing required properties: %s", strings.Join(missingKeys, ", "))
	}
	return nil
}

// Validate checks the body if it is valid and contains all of the required keys
func (b MapBuilder) Validate() error {
	if len(b.requiredKeys) > 0 {
		if err := b.validateRequiredKeys(); err != nil {
			return err
		}
	}

	if b.validateTemplate != "" {
		if err := b.MergeJsonnet(b.validateTemplate, false); err != nil {
			return fmt.Errorf("Validate error. %s", err)
		}
	}

	return nil
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
