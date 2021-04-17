package mapbuilder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/pkg/randdata"
	"github.com/reubenmiller/go-c8y-cli/pkg/timestamp"
	"github.com/tidwall/gjson"
	"go.uber.org/zap/zapcore"
)

const (
	// Separator character which is used when setting the path via a dot notation
	Separator              = "."
	timeFormatRFC3339Micro = "2006-01-02T15:04:05.999Z07:00"
)

var Logger *logger.Logger

func init() {
	Logger = logger.NewLogger("mapbuilder", logger.Options{
		Level:  zapcore.DebugLevel,
		Color:  true,
		Silent: true,
	})
}

func registerNativeFuntions(vm *jsonnet.VM) {
	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Name",
		Params: ast.Identifiers{"prefix", "postfix"},
		Func: func(parameters []interface{}) (interface{}, error) {
			if len(parameters) == 0 {
				return randdata.Name(), nil
			}
			if len(parameters) == 1 {
				return randdata.Name(parameters[0].(string)), nil
			}
			if len(parameters) > 1 {
				return randdata.Name(parameters[0].(string), parameters[1].(string)), nil
			}
			return randdata.Name(), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Password",
		Params: ast.Identifiers{},
		Func: func(parameters []interface{}) (interface{}, error) {
			return randdata.Password(), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Bool",
		Params: ast.Identifiers{},
		Func: func(parameters []interface{}) (interface{}, error) {
			return randdata.Bool(), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Int",
		Params: ast.Identifiers{"maximum"},
		Func: func(parameters []interface{}) (interface{}, error) {
			max := getIntParameter(parameters)
			return float64(randdata.Integer(max)), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Float",
		Params: ast.Identifiers{},
		Func: func(parameters []interface{}) (interface{}, error) {
			return randdata.Float(), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Char",
		Params: ast.Identifiers{"maximum"},
		Func: func(parameters []interface{}) (interface{}, error) {
			max := getIntParameter(parameters)
			return randdata.Char(int(max)), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Digit",
		Params: ast.Identifiers{"maximum"},
		Func: func(parameters []interface{}) (interface{}, error) {
			max := getIntParameter(parameters)
			return randdata.Digit(int(max)), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "AlphaNumeric",
		Params: ast.Identifiers{"maximum"},
		Func: func(parameters []interface{}) (interface{}, error) {
			max := getIntParameter(parameters)
			return randdata.AlphaNumeric(int(max)), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Hex",
		Params: ast.Identifiers{"maximum"},
		Func: func(parameters []interface{}) (interface{}, error) {
			max := getIntParameter(parameters)
			return randdata.Hex(int(max)), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Now",
		Params: ast.Identifiers{"offset"},
		Func: func(parameters []interface{}) (interface{}, error) {

			d, err := timestamp.ParseDurationRelativeToNow(getStringParameter(parameters))
			if err != nil {
				return nil, err
			}
			return d.Format(time.RFC3339), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "NowNano",
		Params: ast.Identifiers{"offset"},
		Func: func(parameters []interface{}) (interface{}, error) {

			d, err := timestamp.ParseDurationRelativeToNow(getStringParameter(parameters))
			if err != nil {
				return nil, err
			}
			return d.Format(time.RFC3339Nano), nil
		},
	})
}

func getIntParameter(parameters []interface{}) int64 {
	if len(parameters) > 0 {
		maximum, err := strconv.ParseInt(fmt.Sprintf("%v", parameters[0]), 10, 64)

		if err != nil {
			return 0
		}
		return maximum
	}
	return 0
}

func getStringParameter(parameters []interface{}) string {
	if len(parameters) > 0 {
		return fmt.Sprintf("%v", parameters[0])
	}
	return ""
}

func evaluateJsonnet(imports string, snippets ...string) (string, error) {
	// Create a JSonnet VM
	vm := jsonnet.MakeVM()
	registerNativeFuntions(vm)

	jsonnetImport := "\n" + `local _ = {Name: std.native("Name"),Password: std.native("Password"),Now: std.native("Now"),NowNano: std.native("NowNano"),Bool: std.native("Bool"),Float: std.native("Float"),Int: std.native("Int"),Hex: std.native("Hex"),Char: std.native("Char"),Digit: std.native("Digit"),AlphaNumeric: std.native("AlphaNumeric")};` + imports

	jsonnetImport += `
// output
`

	if len(snippets) > 0 {
		jsonnetImport += strings.Join(snippets, " +\n")
	} else {
		jsonnetImport += "{}"
	}

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

type IteratorReference struct {
	Path  string
	Value iterator.Iterator
}

// MapBuilder creates body builder
type MapBuilder struct {
	mu                    sync.Mutex
	body                  map[string]interface{}
	bodyIterators         []IteratorReference
	file                  string
	raw                   string
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
		randdata.Password(),
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
					inputImports = append(inputImports, "value: \""+escapeDoubleQuotes(result.String())+"\"")
				}
			}
		}

		if len(inputImports) > 0 {
			localInput = fmt.Sprintf("{%s}", strings.Join(inputImports, ","))
		}
	}

	// add external input to input.value
	externalInput := "{value: null}"
	input = bytes.TrimSpace(input)
	if len(input) > 0 {
		if bytes.HasPrefix(input, []byte("{")) && bytes.HasSuffix(input, []byte("}")) {
			externalInput = "{value: " + string(input) + "}"
		} else {
			externalInput = fmt.Sprintf("{value: \"%s\" }", escapeDoubleQuotes(string(input)))
		}
	}
	Logger.Debugf("externalInput: %s", externalInput)

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

// SetRaw sets the body to a raw string
func (b *MapBuilder) SetRaw(v string) {
	b.raw = v
}

// GetMap returns the body as a map[string]interface{}
func (b *MapBuilder) GetMap() map[string]interface{} {
	return b.body
}

// GetFileContents returns the map contents as a file (only if a file is already set)
func (b *MapBuilder) GetFileContents() *os.File {
	file, err := os.Open(b.file)
	if err != nil {
		Logger.Errorf("failed to open file. %s", err)
		return nil
	}
	return file
}

// HasFile return true if the body is being set from file
func (b *MapBuilder) HasFile() bool {
	return b.file != ""
}

// HasRaw return true if the body is being set from raw data
func (b *MapBuilder) HasRaw() bool {
	return b.raw != ""
}

// GetRaw get raw body
func (b *MapBuilder) GetRaw() string {
	return b.raw
}

// GetFile get the file reference
func (b *MapBuilder) GetFile() (*os.File, error) {
	return os.Open(b.file)
}

// GetBody returns the body as an interface
func (b *MapBuilder) GetBody() (interface{}, error) {
	if b.HasFile() {
		return os.Open(b.file)
	}
	if b.HasRaw() {
		return b.raw, nil
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

	// TODO: deep clone
	// clone body to eliminate side-effects
	bodyClone := NewInitializedMapBuilder()
	for k, v := range b.body {
		bodyClone.body[k] = v
	}

	for _, it := range b.bodyIterators {
		value, input, itErr := it.Value.GetNext()

		Logger.Debugf("body iterator. path=%s, value=%s", it.Path, value)

		if itErr != nil {
			if itErr == io.EOF {
				err = itErr
				return
			}
		} else {
			switch extInput := input.(type) {
			case []byte:
				b.externalInput = extInput
			}
			Logger.Debugf("setting externalInput: %s", b.externalInput)

			if _, ok := bodyClone.body[it.Path]; !ok && len(value) > 0 {
				tempValue := make(map[string]interface{})
				var bErr error
				// Only set non-object values. Complex objects can be referred to via the input.value template variable
				if jsonErr := json.Unmarshal(value, &tempValue); jsonErr != nil {
					Logger.Debugf("Setting value: key=%s, value=%s", it.Path, value)
					bErr = bodyClone.Set(it.Path, fmt.Sprintf("%s", value))
				}

				if bErr != nil {
					break
				}
			}
		}
	}

	Logger.Debugf("Body (pre marshalling). %v", bodyClone)

	body, err = json.Marshal(bodyClone.body)
	if err != nil {
		return
	}

	Logger.Debugf("Body (pre templating)\nbody:\t%s\n\texternalInput:\t%s", body, b.externalInput)

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

	// store iterators seprately so we can itercept the raw value which is otherwise lost during json marshalling
	if it, ok := value.(iterator.Iterator); ok {
		b.bodyIterators = append(b.bodyIterators, IteratorReference{path, it})
		Logger.Debugf("DEBUG: Found iterator. path=%s", path)
		return nil
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

func escapeDoubleQuotes(v string) string {
	return strings.ReplaceAll(v, `"`, `\"`)
}
