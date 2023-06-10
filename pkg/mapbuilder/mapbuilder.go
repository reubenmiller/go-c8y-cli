package mapbuilder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/randdata"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/timestamp"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap/zapcore"
)

const (
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
			prefix := getStringParameter(parameters, 0)
			postfix := getStringParameter(parameters, 1)
			return randdata.Name(prefix, postfix), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Password",
		Params: ast.Identifiers{"length"},
		Func: func(parameters []interface{}) (interface{}, error) {
			length := getIntParameter(parameters, 0)
			return randdata.Password(int(length)), nil
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
		Params: ast.Identifiers{"max", "min"},
		Func: func(parameters []interface{}) (interface{}, error) {
			max := getIntParameter(parameters, 0)
			min := getIntParameter(parameters, 1)
			return float64(randdata.Integer(max, min)), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Float",
		Params: ast.Identifiers{"max", "min", "precision"},
		Func: func(parameters []interface{}) (interface{}, error) {
			max := getFloatParameter(parameters, 0)
			min := getFloatParameter(parameters, 1)
			precision := getIntParameter(parameters, 2)
			return randdata.Float(max, min, int(precision)), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Char",
		Params: ast.Identifiers{"maximum"},
		Func: func(parameters []interface{}) (interface{}, error) {
			max := getIntParameter(parameters, 0)
			return randdata.Char(int(max)), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Digit",
		Params: ast.Identifiers{"maximum"},
		Func: func(parameters []interface{}) (interface{}, error) {
			max := getIntParameter(parameters, 0)
			return randdata.Digit(int(max)), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "AlphaNumeric",
		Params: ast.Identifiers{"maximum"},
		Func: func(parameters []interface{}) (interface{}, error) {
			max := getIntParameter(parameters, 0)
			return randdata.AlphaNumeric(int(max)), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Hex",
		Params: ast.Identifiers{"maximum"},
		Func: func(parameters []interface{}) (interface{}, error) {
			max := getIntParameter(parameters, 0)
			return randdata.Hex(int(max)), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Now",
		Params: ast.Identifiers{"offset"},
		Func: func(parameters []interface{}) (interface{}, error) {

			d, err := timestamp.ParseTimestamp(getStringParameter(parameters, 0))
			if err != nil {
				return nil, err
			}
			RFC3339Milli := "2006-01-02T15:04:05.000Z07:00"
			return d.Format(RFC3339Milli), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "NowNano",
		Params: ast.Identifiers{"offset"},
		Func: func(parameters []interface{}) (interface{}, error) {

			d, err := timestamp.ParseTimestamp(getStringParameter(parameters, 0))
			if err != nil {
				return nil, err
			}
			return d.Format(time.RFC3339Nano), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "GetURLPath",
		Params: ast.Identifiers{"url"},
		Func: func(parameters []interface{}) (interface{}, error) {
			uri := getStringParameter(parameters, 0)
			p, err := url.Parse(uri)
			if err != nil {
				return "", err
			}
			out := p.EscapedPath()
			if p.RawQuery != "" {
				out += "?" + p.RawQuery
			}
			return out, nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "GetURLHost",
		Params: ast.Identifiers{"url"},
		Func: func(parameters []interface{}) (interface{}, error) {
			uri := getStringParameter(parameters, 0)
			p, err := url.Parse(uri)
			if err != nil {
				return "", err
			}
			out := ""
			if p.Host != "" {
				if p.Scheme != "" {
					out = p.Scheme + "://"
				}
				out += p.Host
			}
			return out, nil
		},
	})
}

func getIntParameter(parameters []interface{}, i int) int64 {
	if len(parameters) > 0 && i < len(parameters) {
		maximum, err := strconv.ParseInt(fmt.Sprintf("%v", parameters[i]), 10, 64)

		if err != nil {
			return 0
		}
		return maximum
	}
	return 0
}

func getFloatParameter(parameters []interface{}, i int) float64 {
	if len(parameters) > 0 && i < len(parameters) {
		value, err := strconv.ParseFloat(fmt.Sprintf("%v", parameters[i]), 64)

		if err != nil {
			return 0
		}
		return value
	}
	return 0
}

func getStringParameter(parameters []interface{}, i int) string {
	if len(parameters) > 0 && i < len(parameters) {
		switch v := parameters[i].(type) {
		case float64, float32:
			return fmt.Sprintf("%f", v)
		default:
			return fmt.Sprintf("%v", parameters[i])
		}
	}
	return ""
}

var customJSONNetFunctions = []string{
	`Name: function(prefix='',postfix='') std.native("Name")(prefix,postfix)`,
	`GetURLPath: std.native("GetURLPath")`,
	`GetURLHost: std.native("GetURLHost")`,
	`Password: function(length=32) std.native("Password")(length)`,
	`Now: function(offset='0s') std.native("Now")(std.toString(offset))`,
	`NowNano: function(offset='0s') std.native("NowNano")(std.toString(offset))`,
	`Bool: std.native("Bool")`,
	`Float: function(max=1,min=0,precision=4) std.native("Float")(max,min,precision)`,
	`Int: function(max=100,min=0) std.native("Int")(max,min)`,
	`Hex: function(max=16) std.native("Hex")(max)`,
	`Char: function(max=16) std.native("Char")(max)`,
	`Digit: function(max=16) std.native("Digit")(max)`,
	`AlphaNumeric: function(max=16) std.native("AlphaNumeric")(max)`,
	`StripKeys: function(value={}) value + {lastUpdated::'','self'::'',creationTime::'',additionParents::'',assetParents::'',childAdditions::'',childAssets::'',childDevices::'',deviceParents::''}`,
	`Get: function(key, value={}, defaultValue={}) if std.type(value) == "object" && std.objectHas(value, key) then {[key]: value[key]} else {[key]: defaultValue}`,
	`Merge: function(key, a={}, b={}) _.Get(key, a, if std.type(b) == "array" then [] else {}) + {[key]+: b}`,
}

func evaluateJsonnet(imports string, snippets ...string) (string, error) {
	// Create a JSonnet VM
	vm := jsonnet.MakeVM()
	registerNativeFuntions(vm)

	jsonnetImport := "\n" + "local _ = {" + strings.Join(customJSONNetFunctions, ",") + "};" + imports

	jsonnetImport += `
// output
`

	if len(snippets) > 0 {
		jsonnetImport += strings.Join(snippets, " +\n")
	} else {
		jsonnetImport += "{}"
	}

	debugJsonnet := strings.EqualFold(os.Getenv("C8Y_JSONNET_DEBUG"), "true")
	hideJsonnetHints := strings.EqualFold(os.Getenv("C8Y_JSONNET_HINT"), "false")

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

		helpMsg := ""
		if !hideJsonnetHints {
			helpMsg = "\nHint:\nSome shells are sensitive about double quotes, try escaping any double quotes:\n\n\t--template \"{\\\"name\\\": \\\"my example text\\\"}\"\n\nAlternatively, jsonnet is more relaxed than json so you can use single quotes:\n\n\t--template \"{name: 'my example text'}\"\n"
		}
		err = fmt.Errorf("Could not create json from template. Error: %s%s", err, helpMsg)
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
func NewInitializedMapBuilder(initBody bool) *MapBuilder {
	builder := NewMapBuilder()
	builder.templates = make([]string, 0)
	builder.autoApplyTemplate = true

	if initBody {
		builder.SetEmptyMap()
	}
	return builder
}

// NewMapBuilderWithInit returns a new map builder seeding the builder with json
func NewMapBuilderWithInit(body []byte) *MapBuilder {
	return &MapBuilder{
		BodyRaw: body,
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
	return NewMapBuilderWithInit([]byte(data)), nil
}

type IteratorReference struct {
	Path  string
	Value iterator.Iterator
}

type LocalVariable struct {
	Name  string
	Value any
}

// MapBuilder creates body builder
type MapBuilder struct {
	mu                    sync.Mutex
	BodyRaw               []byte
	bodyOptional          map[string]interface{}
	bodyIterators         []IteratorReference
	file                  string
	raw                   string
	TemplateIterator      iterator.Iterator
	TemplateIteratorNames []string

	// Array output settings
	ArraySize   int
	ArrayPrefix string
	ArraySuffix string

	templateVariables      map[string]interface{}
	requiredKeys           []string
	autoApplyTemplate      bool
	templates              []string
	externalInput          []byte
	localTemplateVariables []LocalVariable
}

func (b *MapBuilder) HasChanged() bool {
	if len(b.templateVariables) > 0 {
		return true
	}

	if len(b.templates) > 0 {
		return true
	}

	if b.file != "" {
		return true
	}

	if b.HasRaw() {
		if b.raw != "" {
			return true
		}
	}

	if len(b.BodyRaw) > 0 {
		return true
		// if !bytes.Equal(b.BodyRaw, []byte("{}")) {
		// }
	}

	if len(b.bodyOptional) > 0 {
		return true
	}

	return false
}

// AppendTemplate appends a templates to be merged in with the body
func (b *MapBuilder) AppendTemplate(template string) *MapBuilder {
	b.templates = append(b.templates, template)
	return b
}

// AddLocalTemplateVariable will inject new local variables into the jsonnet template
// using the local name = value; syntax. It allows users to access additional variables
// more easily
func (b *MapBuilder) AddLocalTemplateVariable(name string, value any) error {
	if b.localTemplateVariables == nil {
		b.localTemplateVariables = make([]LocalVariable, 0, 1)
	}

	switch v := value.(type) {
	case string:
		if json.Valid([]byte(v)) {
			b.localTemplateVariables = append(b.localTemplateVariables, LocalVariable{
				Name:  name,
				Value: string(v),
			})
		} else {
			if outB, err := json.Marshal(v); err == nil {
				b.localTemplateVariables = append(b.localTemplateVariables, LocalVariable{
					Name:  name,
					Value: string(outB),
				})
			}
		}

	case int, int16, int32, int64, float32, float64, bool:
		b.localTemplateVariables = append(b.localTemplateVariables, LocalVariable{
			Name:  name,
			Value: v,
		})
	default:
		jsonV, jsonErr := json.Marshal(v)
		if jsonErr != nil {
			// Fallback to a plain string
			b.localTemplateVariables = append(b.localTemplateVariables, LocalVariable{
				Name:  name,
				Value: fmt.Sprintf("'%s'", strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "\\'")),
			})
		} else {
			b.localTemplateVariables = append(b.localTemplateVariables, LocalVariable{
				Name:  name,
				Value: string(jsonV),
			})
		}

	}
	return nil
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
	b.BodyRaw = []byte("{}")
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

	// Only merge in existing JSON if it is not just an empty object
	// as the other templates might not be objects which can be merged together in jsonne
	// e.g. "_.Int(1) + {}"  will cause an error
	if !bytes.Equal(existingJSON, []byte("{}")) {
		if appendTemplates {
			templates = append([]string{string(existingJSON)}, templates...)
		} else {
			templates = append(templates, string(existingJSON))
		}
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
		randdata.Password(32),
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

	// add custom local variables
	var localVariables strings.Builder
	for _, localVar := range b.localTemplateVariables {
		localVariables.WriteString(fmt.Sprintf("\nlocal %s = %s;", localVar.Name, localVar.Value))
	}

	return fmt.Sprintf("\nlocal vars = %s;\n%s\n%s\n%s\n%s\n%s\n", jsonStr, varsHelper, randomHelper, timeHelper, inputHelper, localVariables.String()), nil
}

// SetMap sets a new map to the body (if not nil). This will remove any existing values in the body
func (b *MapBuilder) SetMap(body map[string]interface{}) {
	if body != nil {
		tempBody, err := json.Marshal(body)
		if err == nil {
			b.BodyRaw = tempBody
		}
	}
}

// ClearMap removed the existing map and sets it to nil.
func (b *MapBuilder) ClearMap() {
	b.BodyRaw = []byte("{}")
}

// ApplyMap sets a new map to the body. This will remove any existing values in the body
func (b *MapBuilder) ApplyMap(body map[string]interface{}) {
	out, err := json.Marshal(body)
	if err != nil {
		Logger.Warningf("Failed to convert map to json. %s", err)
	} else {
		b.BodyRaw = out
	}
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
	out := make(map[string]interface{})
	_ = jsonUtilities.ParseJSON(string(b.BodyRaw), out)
	return out
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
	return gjson.GetBytes(b.BodyRaw, key).Value()
}

// GetString the value as a string
func (b *MapBuilder) GetString(key string) (string, bool) {
	val := gjson.GetBytes(b.BodyRaw, key)
	return val.String(), val.Exists()
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

func (b *MapBuilder) MarshalJSON() (body []byte, err error) {
	if b.ArraySize > 0 {
		return b.MarshalJSONArray()
	}
	return b.MarshalJSONObject()
}

func (b *MapBuilder) MarshalJSONArray() ([]byte, error) {
	out := make([]byte, 0)
	out = append(out, b.ArrayPrefix...)
	var outErr error
	for i := 0; i < b.ArraySize; i++ {

		line, err := b.MarshalJSONObject()

		if err == io.EOF {
			if i == 0 {
				outErr = io.EOF
			}
			break
		}

		if line != nil {
			if i > 0 {
				out = append(out, ',')
			}
			out = append(out, line...)
		}

		if err == io.EOF {
			outErr = err
			break
		}
		if err != nil {
			return nil, err
		}
	}
	out = append(out, b.ArraySuffix...)
	return out, outErr
}

// MarshalJSON returns the body as json
func (b *MapBuilder) MarshalJSONObject() (body []byte, err error) {
	if !b.HasChanged() {
		return nil, nil
	}
	body = []byte(b.BodyRaw)

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

			// NOTE: Do not overwrite existing values if non empty
			if len(value) > 0 {
				if v := gjson.GetBytes(body, it.Path); !v.Exists() {

					// Only assign non-object and non-array values, as these types will be handled in the templating engine
					valueObj := gjson.ParseBytes(value)
					if !(valueObj.IsObject() || valueObj.IsArray()) {
						bodyTemp, bErr := sjson.SetBytes(body, it.Path, value)
						if bErr != nil {
							Logger.Warningf("Could not set bytes. Ignoring value: path=%s, value=%s, err=%", it.Path, value, bErr)
							continue
						}
						body = bodyTemp
					}
				}
			}
		}
	}

	// merge optional values, but prefer already set values
	if b.bodyOptional != nil {
		bodyTemp, bErr := b.MergeJSON(body, b.bodyOptional)
		if bErr != nil {
			return nil, bErr

		}
		body = bodyTemp
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

// KeyExists return true if the given dot notation path exists or not
func (b *MapBuilder) KeyExists(path string) bool {
	return gjson.GetBytes(b.BodyRaw, path).Exists()
}

// SetOptionalMap set optional map values which can be overwritten by other values
func (b *MapBuilder) SetOptionalMap(value map[string]interface{}) {
	b.bodyOptional = value
}

func (b *MapBuilder) SetPath(path string, value interface{}) error {
	out, err := sjson.SetBytes(b.BodyRaw, path, value)
	if err != nil {
		return err
	}
	b.BodyRaw = out
	return nil
}

func (b *MapBuilder) SetRawPath(path string, value []byte) error {
	out, err := sjson.SetRawBytes(b.BodyRaw, path, value)
	if err != nil {
		return err
	}
	b.BodyRaw = out
	return err
}

// Set sets a value to a give dot notation path
func (b *MapBuilder) Set(path string, value interface{}) error {
	// store iterators seprately so we can itercept the raw value which is otherwise lost during json marshalling
	if it, ok := value.(iterator.Iterator); ok {
		b.bodyIterators = append(b.bodyIterators, IteratorReference{path, it})
		Logger.Debugf("DEBUG: Found iterator. path=%s", path)
		return nil
	}

	if err := b.SetPath(path, value); err != nil {
		return err
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

	temp, err := b.MergeJSON(b.BodyRaw, maps...)
	if err != nil {
		return err
	}

	b.BodyRaw = temp
	return nil
}

func (b *MapBuilder) MergeJSON(existingJSON []byte, maps ...map[string]interface{}) ([]byte, error) {
	if len(maps) == 0 {
		return existingJSON, nil
	}

	if len(existingJSON) > 0 {
		tempBody := make(map[string]interface{})
		if err := jsonUtilities.ParseJSON(string(existingJSON), tempBody); err != nil {
			return nil, err
		}
		maps = append(maps, tempBody)
	}

	tempOutput := mergeMaps(maps...)

	return json.Marshal(tempOutput)
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
