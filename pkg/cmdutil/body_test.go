package cmdutil

import (
	"encoding/json"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/assert"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
)

// newSoftwareVersionBody builds a body the same way as "software versions create"
func newSoftwareVersionBody(t *testing.T, template string, flagValues map[string]string) *mapbuilder.MapBuilder {
	t.Helper()
	body := mapbuilder.NewInitializedMapBuilder(true)
	for k, v := range flagValues {
		assert.OK(t, body.Set(k, v))
	}
	body.PrependTemplate(`{type: 'c8y_SoftwareBinary', c8y_Global:{}}`)
	if template != "" {
		body.AppendTemplate(template)
	}
	body.SetRequiredKeys("type")
	return body
}

var softwareDefaults = []BodyDefault{
	{Path: "c8y_Software.version", Value: ""},
	{Path: "c8y_Software.url", Value: ""},
}

// Issue #378: the piped input must be available to the template (input.value)
// and must be re-evaluated for each request
func Test_RenderBodyWithInput_UsesPipedInputPerRequest(t *testing.T) {
	body := newSoftwareVersionBody(t, `{c8y_Software:{version: input.value.version}}`, nil)

	out, err := RenderBodyWithInput(body, []byte(`{"name":"app1","version":"1.0.0"}`), softwareDefaults...)
	assert.OK(t, err)
	assertJSONEqual(t, out, `{"c8y_Global":{},"c8y_Software":{"url":"","version":"1.0.0"},"type":"c8y_SoftwareBinary"}`)

	out, err = RenderBodyWithInput(body, []byte(`{"name":"app2","version":"2.0.0"}`), softwareDefaults...)
	assert.OK(t, err)
	assertJSONEqual(t, out, `{"c8y_Global":{},"c8y_Software":{"url":"","version":"2.0.0"},"type":"c8y_SoftwareBinary"}`)
}

// The defaults must not override values set by a template
func Test_RenderBodyWithInput_DefaultsDoNotOverrideTemplate(t *testing.T) {
	body := newSoftwareVersionBody(t, `{c8y_Software:{version: '9.9', url: 'https://example.com'}}`, nil)
	out, err := RenderBodyWithInput(body, "12345", softwareDefaults...)
	assert.OK(t, err)
	assertJSONEqual(t, out, `{"c8y_Global":{},"c8y_Software":{"url":"https://example.com","version":"9.9"},"type":"c8y_SoftwareBinary"}`)
}

// Without any template or flags, the defaults are still applied (empty version and url)
func Test_RenderBodyWithInput_DefaultsWhenNothingIsSet(t *testing.T) {
	body := newSoftwareVersionBody(t, "", nil)
	out, err := RenderBodyWithInput(body, "12345", softwareDefaults...)
	assert.OK(t, err)
	assertJSONEqual(t, out, `{"c8y_Global":{},"c8y_Software":{"url":"","version":""},"type":"c8y_SoftwareBinary"}`)
}

// Explicit flag values still take precedence, and non piped input (e.g. a --software value) gives input.value = null
func Test_RenderBodyWithInput_FlagValuesWithNonPipedInput(t *testing.T) {
	body := newSoftwareVersionBody(t, `{extra: input.value}`, map[string]string{
		"c8y_Software.version": "1.2.3",
		"c8y_Software.url":     "https://example.com",
	})
	out, err := RenderBodyWithInput(body, "12345", softwareDefaults...)
	assert.OK(t, err)
	assertJSONEqual(t, out, `{"c8y_Global":{},"c8y_Software":{"url":"https://example.com","version":"1.2.3"},"extra":null,"type":"c8y_SoftwareBinary"}`)
}

func Test_SetBodyValue(t *testing.T) {
	out, err := SetBodyValue([]byte(`{"c8y_Software":{"version":"1.0.0"}}`), "c8y_Software.url", "https://example.com/binary")
	assert.OK(t, err)
	assertJSONEqual(t, out, `{"c8y_Software":{"url":"https://example.com/binary","version":"1.0.0"}}`)
}

// assertJSONEqual compares json ignoring the key order
func assertJSONEqual(t *testing.T, got []byte, want string) {
	t.Helper()
	var gotV, wantV any
	assert.OK(t, json.Unmarshal(got, &gotV))
	assert.OK(t, json.Unmarshal([]byte(want), &wantV))
	assert.EqualMarshalJSON(t, gotV, mustMarshal(t, wantV))
}

func mustMarshal(t *testing.T, v any) string {
	t.Helper()
	out, err := json.Marshal(v)
	assert.OK(t, err)
	return string(out)
}
