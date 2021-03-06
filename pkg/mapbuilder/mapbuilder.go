package mapbuilder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/sethvargo/go-password/password"
	"github.com/tidwall/gjson"
)

const (
	// Separator character which is used when setting the path via a dot notation
	Separator              = "."
	timeFormatRFC3339Micro = "2006-01-02T15:04:05.999Z07:00"
)

func evaluateJsonnet(imports string, snippets ...string) (string, error) {
	// Create a JSonnet VM
	vm := jsonnet.MakeVM()

	jsonnetImport := imports

	jsonnetImport += `
// body
`

	// base = strings.TrimSpace(base)

	// // Add closing ";" if required
	// if strings.HasSuffix(base, "}") {
	// 	base = base + ";"
	// }

	// jsonnetImport += "\nlocal base = " + base

	// jsonnetImport += "\nlocal final = "
	if len(snippets) > 0 {
		jsonnetImport += strings.Join(snippets, " +\n")
	} else {
		jsonnetImport += "{}"
	}

	// jsonnetImport += "\nfinal"

	debugJsonnet := strings.EqualFold(os.Getenv("C8Y_JSONNET_DEBUG"), "true")

	if debugJsonnet {
		log.Printf("jsonnet template: %s\n", jsonnetImport)
	}

	// evaluate the jsonnet
	out, err := vm.EvaluateAnonymousSnippet("file", jsonnetImport)

	if err != nil {

		if debugJsonnet {
			// Include full template (with injected variables/functions) otherwise the error
			// will report line numbers that the user does not know about
			log.Printf("jsonnet error: %s", err)
		}

		err = fmt.Errorf("Could not create json from template. Error: %s", err)
	}
	return out, err
}

// NewMapBuilder creates a new map builder with the map set to nil
func NewMapBuilder() *MapBuilder {
	return &MapBuilder{
		templates:         []string{},
		autoApplyTemplate: true,
	}
}

// NewInitializedMapBuilder creates a new map builder with the map set to an empty map
func NewInitializedMapBuilder() *MapBuilder {
	builder := NewMapBuilder()
	builder.templates = make([]string, 0)
	builder.autoApplyTemplate = true
	builder.SetEmptyMap()
	return builder
}

// NewMapBuilderWithInit returns a new map builder seeding the builder with the give map
func NewMapBuilderWithInit(body map[string]interface{}) *MapBuilder {
	return &MapBuilder{
		body: body,
	}
}

// NewMapBuilderFromJsonnetSnippet returns a new mapper builder from a jsonnet snippet
// See https://jsonnet.org/learning/tutorial.html for details on how to create a snippet
func NewMapBuilderFromJsonnetSnippet(snippet string) (*MapBuilder, error) {
	jsonStr, err := evaluateJsonnet("", snippet)

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
	mu                    sync.Mutex
	body                  map[string]interface{}
	file                  string
	TemplateIterator      iterator.Iterator
	TemplateIteratorNames []string

	templateVariables map[string]interface{}
	requiredKeys      []string
	autoApplyTemplate bool
	templates         []string
	externalInput     []byte
}

// AppendTemplate appends a templates to be merged in with the body
func (b *MapBuilder) AppendTemplate(template string) *MapBuilder {
	b.templates = append(b.templates, template)
	return b
}

// PrependTemplate prepends a templates to be merged in with the body
func (b *MapBuilder) PrependTemplate(template string) *MapBuilder {
	b.templates = append([]string{template}, b.templates[:]...)
	return b
}

// SetApplyTemplateOnMarshalPreference sets whether the templates should be applied during marshalling or not.
func (b *MapBuilder) SetApplyTemplateOnMarshalPreference(value bool) *MapBuilder {
	b.autoApplyTemplate = value
	return b
}

// SetEmptyMap sets the body to an empty map. It will override an existing body
func (b *MapBuilder) SetEmptyMap() *MapBuilder {
	b.body = make(map[string]interface{})
	return b
}

// ApplyTemplates merges the existing body data with a given jsonnet snippet.
// When reverse is false, then the snippet will be applied to the existing data,
// when reverse is true, then the given snippet will be the base, and the existing data will be applied to the new snippet.
func (b *MapBuilder) ApplyTemplates(existingJSON []byte, input []byte, appendTemplates bool) ([]byte, error) {
	var err error
	if len(existingJSON) == 0 {
		existingJSON = []byte("{}")
	}
	existingJSON = bytes.TrimSpace(existingJSON)

	imports, err := b.getTemplateVariablesJsonnet(existingJSON, input)
	if err != nil {
		return nil, err
	}
	templates := []string{}
	for _, template := range b.templates {
		templates = append(templates, strings.TrimSpace(template))
	}

	if appendTemplates {
		templates = append([]string{string(existingJSON)}, templates...)
	} else {
		templates = append(templates, string(existingJSON))
	}

	var mergedJSON string
	mergedJSON, err = evaluateJsonnet(imports, templates...)

	if err != nil {
		return nil, fmt.Errorf("failed to merge json. %w", err)
	}

	return []byte(mergedJSON), nil
}

// SetTemplateVariables stores the given variables that will be used in the template evaluation
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

func (b *MapBuilder) getTemplateVariablesJsonnet(existingJSON []byte, input []byte) (string, error) {
	jsonStr := []byte("{}")
	// default to empty object (if no custom template variables are provided)
	if b.templateVariables != nil {
		v, err := json.Marshal(b.templateVariables)
		if err != nil {
			return "", err
		}
		jsonStr = v
	}

	varsHelper := `local var(prop, defaultValue="") = if std.objectHas(vars, prop) then vars[prop] else defaultValue;`

	// Seed random otherwise it will not change with execution
	rand.Seed(time.Now().UTC().UnixNano())

	randomHelper := fmt.Sprintf(`local rand = { bool: %t, int: %d, int2: %d, float: %f, float2: %f, float3: %f, float4: %f, password: "%s" };`,
		rand.Float32() > 0.5,
		rand.Intn(100),
		rand.Intn(100),
		rand.Float32(),
		rand.Float32(),
		rand.Float32(),
		rand.Float32(),
		generatePassword(),
	)

	index := "1"

	if b.TemplateIterator != nil {
		nextIndex, _, err := b.TemplateIterator.GetNext()
		if err != nil {
			return "", err
		}
		index = string(nextIndex)
	}

	indexInt := 1
	if v, err := strconv.Atoi(index); err == nil {
		indexInt = v
	}

	localInput := "{}"
	if len(b.TemplateIteratorNames) > 0 {
		inputImports := []string{}

		results := gjson.GetManyBytes(existingJSON, b.TemplateIteratorNames...)

		for i, result := range results {
			if result.Exists() {
				if result.IsObject() {
					inputImports = append(inputImports, b.TemplateIteratorNames[i]+": '"+result.String()+"'")
				} else {
					inputImports = append(inputImports, "value: '"+result.String()+"'")
				}
			}
		}

		if len(inputImports) > 0 {
			localInput = fmt.Sprintf("{%s}", strings.Join(inputImports, ","))
		}
	}

	// add external input to input.value
	externalInput := "{}"
	input = bytes.TrimSpace(input)
	if len(input) > 0 && bytes.HasPrefix(input, []byte("{")) && bytes.HasSuffix(input, []byte("}")) {
		externalInput = "{value: " + string(input) + "}"
	}

	inputHelper := fmt.Sprintf(`local input = {index: %d} + %s + %s;`,
		indexInt,
		localInput,
		externalInput,
	)

	timeHelper := fmt.Sprintf(`local time = {now: "%s", nowNano: "%s"};`,
		time.Now().Format(timeFormatRFC3339Micro),
		time.Now().Format(time.RFC3339Nano),
	)
	return fmt.Sprintf("\nlocal vars = %s;\n%s\n%s\n%s\n%s\n", jsonStr, varsHelper, randomHelper, timeHelper, inputHelper), nil
}

// SetMap sets a new map to the body (if not nil). This will remove any existing values in the body
func (b *MapBuilder) SetMap(body map[string]interface{}) {
	if body != nil {
		b.body = body
	}
}

// ClearMap removed the existing map and sets it to nil.
func (b *MapBuilder) ClearMap() {
	b.body = nil
}

// ApplyMap sets a new map to the body. This will remove any existing values in the body
func (b *MapBuilder) ApplyMap(body map[string]interface{}) {
	b.body = body
}

// SetFile sets the body to the contents of the file path
func (b *MapBuilder) SetFile(path string) {
	b.file = path
}

// GetMap returns the body as a map[string]interface{}
func (b *MapBuilder) GetMap() map[string]interface{} {
	return b.body
}

// GetFileContents returns the map contents as a file (only if a file is already set)
func (b *MapBuilder) GetFileContents() *os.File {
	file, err := os.Open(b.file)
	if err != nil {
		log.Printf("failed to open file. %s", err)
		return nil
	}
	return file
}

// HasFile return true if the body is being set from file
func (b *MapBuilder) HasFile() bool {
	return b.file != ""
}

// GetBody returns the body as an interface
func (b *MapBuilder) GetBody() (interface{}, error) {
	if b.HasFile() {
		return os.Open(b.file)
	}
	return b.GetMap(), nil
}

// Get returns a value as an interface
func (b *MapBuilder) Get(key string) interface{} {
	return b.body[key]
}

// GetString the value as a string
func (b *MapBuilder) GetString(key string) (string, bool) {
	val, ok := b.body[key].(string)
	return val, ok
}

// SetRequiredKeys stores the list of keys which should be present when marshaling the map to json.
// Nested paths are accepted via dot notation
func (b *MapBuilder) SetRequiredKeys(keys ...string) {
	b.requiredKeys = keys
}

func (b *MapBuilder) validateRequiredKeysBytes(body []byte) error {
	missingKeys := make([]string, 0)
	results := gjson.GetManyBytes(body, b.requiredKeys...)

	for i, path := range results {
		if !path.Exists() {
			missingKeys = append(missingKeys, b.requiredKeys[i])
		}
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf("Body is missing required properties: %s", strings.Join(missingKeys, ", "))
	}
	return nil
}

// MarshalJSONWithInput convers the body to json and also injecting additional data into the template input to make
// it available using the input.value variable in jsonnet
func (b *MapBuilder) MarshalJSONWithInput(input interface{}) (body []byte, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch v := input.(type) {
	case []byte:
		b.externalInput = v
	case string:
		b.externalInput = []byte(v)
	}
	return b.MarshalJSON()
}

// MarshalJSON returns the body as json
func (b *MapBuilder) MarshalJSON() (body []byte, err error) {
	if b.body == nil {
		// should return empty object? or add as option?
		body = []byte("{}")
		return
	}

	// evaluate any iterators
	body, err = json.Marshal(b.body)
	if err != nil {
		return
	}

	if b.autoApplyTemplate && len(b.templates) > 0 {
		body, err = b.ApplyTemplates(body, b.externalInput, false)
		if err != nil {
			return
		}
	}

	// Validate after applying the template
	if err := b.validateRequiredKeysBytes(body); err != nil {
		return nil, err
	}
	return
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

// MergeMaps merges a list of maps into the body. If the body does not already exists,
// then it will be ignored. Only shallow merging is done.
// Duplicate keys will be overwritten by maps later in the list
func (b *MapBuilder) MergeMaps(maps ...map[string]interface{}) error {
	if len(maps) == 0 {
		return nil
	}

	if b.body != nil {
		maps = append([]map[string]interface{}{b.body}, maps...)
	}

	b.body = mergeMaps(maps...)
	return nil
}

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
