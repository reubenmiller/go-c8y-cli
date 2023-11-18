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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flatten"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonfilter"
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
		Name:   "Date",
		Params: ast.Identifiers{"now", "offset", "format", "utc"},
		Func: func(parameters []interface{}) (interface{}, error) {
			offsetRaw := getParameter(parameters, 1)
			offset := ""
			switch value := offsetRaw.(type) {
			case string:
				offset = value
			default:
				offset = fmt.Sprintf("%vs", value)
			}

			var d time.Time
			var err error

			nowRaw := getParameter(parameters, 0)
			if nowRaw == nil {
				return nil, fmt.Errorf("missing date")
			}

			switch now := nowRaw.(type) {
			case float64:
				d, err = timestamp.AddDateTime(fmt.Sprintf("%v", int(now)), offset)
			case float32:
				d, err = timestamp.AddDateTime(fmt.Sprintf("%v", int(now)), offset)
			case string:
				d, err = timestamp.AddDateTime(now, offset)
			default:
				return nil, fmt.Errorf("invalid date format")
			}
			if err != nil {
				return nil, err
			}

			format := "2006-01-02T15:04:05.000Z07:00"
			if timeFormat := strings.TrimSpace(getStringParameter(parameters, 2)); timeFormat != "" {
				format = timeFormat
			}

			useUTC := getBooleanParameter(parameters, 3, false)
			if useUTC {
				return d.UTC().Format(format), nil
			}

			return d.Format(format), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Duration",
		Params: ast.Identifiers{"dateA", "dateB", "unit"},
		Func: func(parameters []interface{}) (interface{}, error) {
			// duration = DateA - DateB
			dateA := getStringParameter(parameters, 0)
			dateB := getStringParameter(parameters, 1)

			tsA, err := timestamp.ParseTimestamp(dateA)
			if err != nil {
				return nil, err
			}

			tsB, err := timestamp.ParseTimestamp(dateB)
			if err != nil {
				return nil, err
			}

			diff := tsA.Sub(tsB)

			// Round date to nearest milliseconds
			diff = diff.Round(time.Millisecond)

			unit := getStringParameter(parameters, 2)

			switch unit {
			case "string", "str", "duration":
				return diff.String(), nil
			case "milliseconds", "ms":
				return float64(diff.Milliseconds()), nil
			case "days", "d":
				return diff.Hours() / 24, nil
			case "hours", "h", "hrs", "hr":
				return diff.Hours(), nil
			case "minutes", "mins", "m":
				return diff.Minutes(), nil
			case "seconds", "s", "sec":
				return diff.Seconds(), nil
			case "object":
				duration := map[string]any{}
				duration["milliseconds"] = float64(diff.Milliseconds())
				duration["seconds"] = diff.Seconds()
				duration["minutes"] = diff.Minutes()
				duration["hours"] = diff.Hours()
				duration["days"] = diff.Hours() / 24
				duration["duration"] = diff.String()
				return duration, nil
			default:
				return diff.String(), nil
			}
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

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "ReplacePattern",
		Params: ast.Identifiers{"value", "from", "to"},
		Func: func(parameters []interface{}) (interface{}, error) {
			value := getStringParameter(parameters, 0)
			from := getStringParameter(parameters, 1)
			to := getStringParameter(parameters, 2)

			pattern, err := regexp.Compile(from)
			if err != nil {
				return "", err
			}
			return pattern.ReplaceAllString(value, to), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "Select",
		Params: ast.Identifiers{"o", "properties"},
		Func: func(parameters []interface{}) (interface{}, error) {
			// Provide a function which uses the same style as the
			// global --select flag
			if len(parameters) < 2 {
				return nil, fmt.Errorf("requires 2 arguments")
			}

			jsonB, err := json.Marshal(parameters[0])
			if err != nil {
				return nil, err
			}
			patterns := []string{}
			switch properties := parameters[1].(type) {
			case []any:
				for _, v := range properties {
					patterns = append(patterns, fmt.Sprintf("%v", v))
				}
			case string:
				// Support people providing a csv string list
				for _, v := range strings.Split(properties, ",") {
					patterns = append(patterns, strings.TrimSpace(v))
				}
			}

			if len(patterns) == 0 {
				return parameters[0], fmt.Errorf("no select values provided")
			}

			flatMap, flatKeys, err := jsonfilter.FilterPropertyByWildcard(string(jsonB), "", patterns, false)
			if err != nil {
				return nil, err
			}
			outB, err := flatten.UnflattenOrdered(flatMap, flatKeys)
			if err != nil {
				return nil, err
			}

			var out any
			if err := json.Unmarshal(outB, &out); err != nil {
				return nil, err
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

func getBooleanParameter(parameters []interface{}, i int, defaultValue bool) bool {
	if len(parameters) > 0 && i < len(parameters) {
		value, err := strconv.ParseBool(fmt.Sprintf("%v", parameters[i]))

		if err != nil {
			return defaultValue
		}
		return value
	}
	return defaultValue
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

func getParameter(parameters []interface{}, i int) any {
	if len(parameters) > 0 && i < len(parameters) {
		return parameters[i]
	}
	return nil
}

func evaluateJsonnet(imports string, snippets ...string) (string, error) {
	// Create a JSonnet VM
	vm := jsonnet.MakeVM()
	registerNativeFuntions(vm)

	// Add functions via jsonnet object
	localFunctions := heredoc.Doc(`
	{
		Name(prefix='',postfix=''):: std.native("Name")(prefix,postfix),
		GetURLPath(url):: std.native("GetURLPath")(url),
		GetURLHost(url):: std.native("GetURLHost")(url),
		Password(length=32):: std.native("Password")(length),
		Now(offset='0s'):: std.native("Now")(std.toString(offset)),
		NowNano(offset='0s'):: std.native("NowNano")(std.toString(offset)),
		Bool():: std.native("Bool"),
		Float(max=1,min=0,precision=4):: std.native("Float")(max,min,precision),
		Int(max=100,min=0):: std.native("Int")(max,min),
		Hex(max=16):: std.native("Hex")(max),
		Char(max=16):: std.native("Char")(max),
		Digit(max=16):: std.native("Digit")(max),
		AlphaNumeric(max=16):: std.native("AlphaNumeric")(max),
		StripKeys(value={}):: value + {lastUpdated::'','self'::'',creationTime::'',additionParents::'',assetParents::'',childAdditions::'',childAssets::'',childDevices::'',deviceParents::''},
		# Deprecated: DeprecatedMerge=>SelectMerge and DeprecatedGet => Select
		DeprecatedMerge(key, a={}, b={}):: _.DeprecatedGet(key, a, if std.type(b) == "array" then [] else {}) + {[key]+: b},
		DeprecatedGet(key, o={}, defaultValue={}):: if std.type(o) == "object" && std.objectHas(o, key) then {[key]: o[key]} else {[key]: defaultValue},
		Date(now="0s", offset="0s", format="", utc=false):: std.native("Date")(now=now, offset=offset, format=format, utc=utc),
		Duration(dateA="", dateB="", unit='object'):: std.native("Duration")(dateA=dateA, dateB=dateB, unit=unit),
		Patch(target={}, patch)::
			local _target = {
				[item.key]: target[item.key]
				for item in std.objectKeysValues(patch)
				if std.objectHas(target, item.key)
			};
			std.mergePatch(_target, patch),
		SelectMerge(a={}, b={})::
			local _keys = std.objectFields(b);
			if std.length(_keys) == 0 then
				{}
			else
				local item = {
					[key]: a[key]
					for key in _keys
					if std.objectHas(a, key)
				};
				item + {
					[key]+: b[key]
					for key in _keys
					if std.isObject(b[key]) || std.isArray(b[key])
				} + {
					[key]: b[key]
					for key in _keys
					if !(std.isObject(b[key]) || std.isArray(b[key]))
				},
		Get(o, f, default=null)::
			local get_(o, ks) =
				if ! std.objectHas(o, ks[0]) then
					default
				else if std.length(ks) == 1 then
					o[ks[0]]
				else
					get_(o[ks[0]], ks[1:]);
				get_(o, std.split(f, '.')),
		Has(o, f)::
			local has_(o, ks) =
				if ! std.objectHas(o, ks[0]) then
					false
				else if std.length(ks) == 1 then
					true
				else
					has_(o[ks[0]], ks[1:]);
			has_(o, std.split(f, '.')),
		RecurseReplace(any, from, to)::
			local recurseReplace_(any, from, to) = (
				{
				object: function(x) { [k]: recurseReplace_(x[k], from, to) for k in std.objectFields(x) },
				array: function(x) [recurseReplace_(e, from, to) for e in x],
				string: function(x) std.native('ReplacePattern')(x, from, to),
				#string: function(x) std.strReplace(x, from, to),
				number: function(x) x,
				boolean: function(x) x,
				'function': function(x) x,
				'null': function(x) x,
				}[std.type(any)](any)
			);
			recurseReplace_(any, from, to),
		ReplacePattern(x, from, to):: std.native('ReplacePattern')(x, from, to),
		Select(o, properties=['*']):: std.native('Select')(o, properties),
	}`)

	jsonnetImport := fmt.Sprintf("\nlocal _ = %s; %s", localFunctions, imports)

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
	log.Printf("Executing output. %s", out)
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
		BodyRaw:           body,
		autoApplyTemplate: true,
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
	appendTemplate         bool
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

// SetAppendTemplatePreference sets if the templates should be appended when evaluating the entire map
// Setting it to append means that the template can override other values provided in the map
func (b *MapBuilder) SetAppendTemplatePreference(value bool) *MapBuilder {
	b.appendTemplate = value
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

type DelayedReader struct {
	Map       *MapBuilder
	Input     any
	readIndex int64
}

func (d *DelayedReader) Read(p []byte) (n int, err error) {
	if d.readIndex > 0 {
		err = io.EOF
		return
	}
	var body []byte
	body, err = d.Map.MarshalJSONWithInput(d.Input)
	n = copy(p, body)
	d.readIndex += int64(n)
	return
}

func (b *MapBuilder) GetDelayedReader(input interface{}) func() io.Reader {
	return func() io.Reader {
		return &DelayedReader{
			Map:   b,
			Input: input,
		}
	}
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
		body, err = b.ApplyTemplates(body, b.externalInput, b.appendTemplate)
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
	// store iterators separately so we can intercept the raw value which is otherwise lost during json marshalling
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

// Set multiple values
func (b *MapBuilder) SetTuple(path string, values ...interface{}) error {
	for _, value := range values {
		// store iterators separately so we can intercept the raw value which is otherwise lost during json marshalling
		if it, ok := value.(iterator.Iterator); ok {
			b.bodyIterators = append(b.bodyIterators, IteratorReference{path, it})
			Logger.Debugf("DEBUG: Found iterator. path=%s", path)
			return nil
		}

		if err := b.SetPath(path, value); err != nil {
			return err
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
