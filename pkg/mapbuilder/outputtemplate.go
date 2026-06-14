package mapbuilder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/randdata"
	sdktemplate "github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/template"
)

// outputTemplateHeader binds the same template environment as the legacy
// evaluateJsonnet path (the `_` helper library, vars/var, rand, time, input
// and the request/response/flags variables). The per-evaluation values are
// read from external variables so the template only has to be compiled once
// and can then be evaluated per row.
func outputTemplateHeader() string {
	return fmt.Sprintf(`
local _ = %s;
local vars = {};
local var(prop, defaultValue="") = if std.objectHas(vars, prop) then vars[prop] else defaultValue;
local rand = std.extVar('rand');
local time = std.extVar('time');
local input = std.extVar('input');
local flags = std.extVar('flags');
local request = std.extVar('request');
local response = std.extVar('response');
`, localFunctions)
}

// OutputTemplate evaluates an --outputTemplate jsonnet snippet. Unlike the
// MapBuilder template path — which re-parses the template and helper library
// on every evaluation — the template is compiled to an AST once and reused,
// which matters when an output template runs once per piped input row.
type OutputTemplate struct {
	mu   sync.Mutex
	eval *sdktemplate.Evaluator
}

// NewOutputTemplate compiles an output template.
func NewOutputTemplate(snippet string) (*OutputTemplate, error) {
	eval, err := sdktemplate.NewEvaluator(snippet, outputTemplateHeader(), registerNativeFunctions)
	if err != nil {
		return nil, formatOutputTemplateError(err)
	}
	return &OutputTemplate{eval: eval}, nil
}

var (
	outputTemplateCacheMu sync.Mutex
	outputTemplateCache   = make(map[string]*OutputTemplate)
)

// GetOutputTemplate returns a compiled output template, reusing a previously
// compiled instance for the same snippet (templates are constant within a
// command invocation, but are evaluated once per piped input row).
func GetOutputTemplate(snippet string) (*OutputTemplate, error) {
	outputTemplateCacheMu.Lock()
	defer outputTemplateCacheMu.Unlock()
	if t, ok := outputTemplateCache[snippet]; ok {
		return t, nil
	}
	t, err := NewOutputTemplate(snippet)
	if err != nil {
		return nil, err
	}
	outputTemplateCache[snippet] = t
	return t, nil
}

// Evaluate runs the template against the output document. input is the piped
// input item associated with the output ([]byte and string values are exposed
// as input.value, anything else as null, matching MarshalJSONWithInput).
// requestData/responseData/flags are bound as the request/response/flags
// template variables.
func (t *OutputTemplate) Evaluate(output []byte, input any, requestData, responseData map[string]any, flags map[string]string) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.eval.SetCode("rand", randJsonnet())
	t.eval.SetCode("time", timeJsonnet())
	t.eval.SetCode("input", inputJsonnet(input))
	t.eval.SetCode("flags", marshalJsonnetValue(flags))
	t.eval.SetCode("request", marshalJsonnetValue(requestData))
	t.eval.SetCode("response", marshalJsonnetValue(responseData))

	out, err := t.eval.Evaluate(output)
	if err != nil {
		return nil, formatOutputTemplateError(err)
	}
	return []byte(out), nil
}

// randJsonnet builds the per-evaluation rand variable using the same value
// construction as getTemplateVariablesJsonnet.
func randJsonnet() string {
	return fmt.Sprintf(`{ bool: %t, int: %d, int2: %d, float: %f, float2: %f, float3: %f, float4: %f, password: "%s" }`,
		rand.Float32() > 0.5,
		rand.Intn(100),
		rand.Intn(100),
		rand.Float32(),
		rand.Float32(),
		rand.Float32(),
		rand.Float32(),
		randdata.Password(32),
	)
}

func timeJsonnet() string {
	return fmt.Sprintf(`{now: "%s", nowNano: "%s"}`,
		time.Now().Format(timeFormatRFC3339Micro),
		time.Now().Format(time.RFC3339Nano),
	)
}

// inputJsonnet builds the input variable with the same shape as
// getTemplateVariablesJsonnet for output templates: a fixed index of 1, and
// the external input bound to input.value (objects as values, anything else
// as a string, no input as null).
func inputJsonnet(input any) string {
	var externalInputBytes []byte
	switch v := input.(type) {
	case []byte:
		externalInputBytes = v
	case string:
		externalInputBytes = []byte(v)
	}

	externalInput := "{value: null}"
	externalInputBytes = bytes.TrimSpace(externalInputBytes)
	if len(externalInputBytes) > 0 {
		if bytes.HasPrefix(externalInputBytes, []byte("{")) && bytes.HasSuffix(externalInputBytes, []byte("}")) {
			externalInput = "{value: " + string(externalInputBytes) + "}"
		} else {
			externalInput = fmt.Sprintf("{value: \"%s\" }", escapeDoubleQuotes(string(externalInputBytes)))
		}
	}
	return "{index: 1} + {} + " + externalInput
}

// marshalJsonnetValue marshals a Go value to json for use as a jsonnet
// external variable, using the same fallbacks as AddLocalTemplateVariable.
func marshalJsonnetValue(value any) string {
	jsonV, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("'%s'", strings.ReplaceAll(fmt.Sprintf("%v", value), "'", "\\'"))
	}
	return string(jsonV)
}

// formatOutputTemplateError reproduces the error formatting of the legacy
// evaluateJsonnet/ApplyTemplates path so template failures read the same.
func formatOutputTemplateError(err error) error {
	debugJsonnet := strings.EqualFold(os.Getenv("C8Y_JSONNET_DEBUG"), "true")
	hideJsonnetHints := strings.EqualFold(os.Getenv("C8Y_JSONNET_HINT"), "false")

	if debugJsonnet {
		log.Printf("jsonnet error: %s", err)
	}

	helpMsg := ""
	if !hideJsonnetHints {
		helpMsg = "\nHint:\nSome shells are sensitive about double quotes, try escaping any double quotes:\n\n\t--template \"{\\\"name\\\": \\\"my example text\\\"}\"\n\nAlternatively, jsonnet is more relaxed than json so you can use single quotes:\n\n\t--template \"{name: 'my example text'}\"\n"
	}
	return fmt.Errorf("failed to merge json. Could not create json from template. Error: %s%s", err, helpMsg)
}
