package mapbuilder

import (
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/assert"
	"github.com/tidwall/pretty"
)

func TestOutputTemplateBindsEnvironment(t *testing.T) {
	tmpl, err := NewOutputTemplate(`{
		name: output.name,
		input: input.value,
		index: input.index,
		method: request.method,
		status: response.statusCode,
		id: flags.id,
		get: _.Get(output, 'nested.value', 'fallback'),
		hasTime: std.length(time.now) > 0,
		randOk: rand.int >= 0,
	}`)
	assert.OK(t, err)

	out, err := tmpl.Evaluate(
		[]byte(`{"name":"dev01","nested":{"value":42}}`),
		"123",
		map[string]any{"method": "GET"},
		map[string]any{"statusCode": 200},
		map[string]string{"id": "agent01"},
	)
	assert.OK(t, err)
	// jsonnet serializes object keys alphabetically
	assert.EqualJSON(t, pretty.Ugly(out), `{"get":42,"hasTime":true,"id":"agent01","index":1,"input":"123","method":"GET","name":"dev01","randOk":true,"status":200}`)
}

func TestOutputTemplateInputValue(t *testing.T) {
	tmpl, err := NewOutputTemplate(`input.value`)
	assert.OK(t, err)

	// no input → null
	out, err := tmpl.Evaluate([]byte(`{}`), nil, nil, nil, nil)
	assert.OK(t, err)
	assert.True(t, strings.TrimSpace(string(out)) == "null")

	// object input is bound as a value
	out, err = tmpl.Evaluate([]byte(`{}`), []byte(`{"id":"1"}`), nil, nil, nil)
	assert.OK(t, err)
	assert.EqualJSON(t, pretty.Ugly(out), `{"id":"1"}`)

	// scalar input is bound as a string
	out, err = tmpl.Evaluate([]byte(`{}`), "55", nil, nil, nil)
	assert.OK(t, err)
	assert.True(t, strings.TrimSpace(string(out)) == `"55"`)
}

func TestOutputTemplateNonJSONOutputBoundAsString(t *testing.T) {
	tmpl, err := NewOutputTemplate(`std.asciiUpper(output)`)
	assert.OK(t, err)

	out, err := tmpl.Evaluate([]byte("plain text"), nil, nil, nil, nil)
	assert.OK(t, err)
	assert.True(t, strings.TrimSpace(string(out)) == `"PLAIN TEXT"`)
}

func TestGetOutputTemplateReusesCompiledTemplates(t *testing.T) {
	a, err := GetOutputTemplate(`{n: 1}`)
	assert.OK(t, err)
	b, err := GetOutputTemplate(`{n: 1}`)
	assert.OK(t, err)
	if a != b {
		t.Errorf("expected the same compiled template instance to be reused")
	}
}

func TestOutputTemplateEvaluationError(t *testing.T) {
	tmpl, err := NewOutputTemplate(`{bad: output.missing.deep}`)
	assert.OK(t, err)

	_, err = tmpl.Evaluate([]byte(`{"a":1}`), nil, nil, nil, nil)
	if err == nil {
		t.Fatal("expected an evaluation error")
	}
	if !strings.Contains(err.Error(), "failed to merge json. Could not create json from template.") {
		t.Errorf("error must keep the legacy formatting, got: %s", err)
	}
	if !strings.Contains(err.Error(), "RUNTIME ERROR: Field does not exist: missing") {
		t.Errorf("error must include the jsonnet runtime error, got: %s", err)
	}
}
