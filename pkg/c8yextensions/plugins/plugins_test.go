package plugins

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const testApp = `{
	"id": "123",
	"name": "devicemanagement-copy",
	"config": {
		"excludedRemotes": {"devicemanagement": ["LWM2Module"]},
		"unknownField": {"nested": [1, 2, 3]},
		"remotes": {
			"devicemanagement": ["DeviceManagementModule"],
			"c8y-asm-ui@3.0.1": ["asm"],
			"other-plugin@1.0.0": ["otherModule"]
		}
	}
}`

const testPlugin = `{
	"name": "my-plugin",
	"contextPath": "my-plugin",
	"manifest": {"remotes": {"my-plugin": ["myModule"]}},
	"applicationVersions": [
		{"version": "1.0.0", "tags": []},
		{"version": "2.0.0", "tags": ["latest"]}
	]
}`

// existing mimics the parsing done in buildBody for a given app
func existing(app gjson.Result) (map[string]ExtensionReference, map[string]interface{}) {
	remotes := map[string]ExtensionReference{}
	preserved := map[string]interface{}{}
	app.Get("config.remotes").ForEach(func(key, value gjson.Result) bool {
		name, version, valid := splitKey(key.String())
		if !valid {
			var v interface{}
			_ = json.Unmarshal([]byte(value.Raw), &v)
			preserved[key.String()] = v
			return true
		}
		ref := ExtensionReference{ContextPath: name, Version: version, Modules: []string{}}
		for _, m := range value.Array() {
			ref.Modules = append(ref.Modules, m.String())
		}
		remotes[name] = ref
		return true
	})
	return remotes, preserved
}

func splitKey(k string) (string, string, bool) {
	for i := 0; i < len(k); i++ {
		if k[i] == '@' {
			return k[:i], k[i+1:], true
		}
	}
	return k, "", false
}

func build(t *testing.T, remotes map[string]ExtensionReference, preserved map[string]interface{}) gjson.Result {
	out, err := buildApplicationBody(gjson.Parse(testApp), remotes, preserved, []byte("{}"))
	require.NoError(t, err)
	return gjson.ParseBytes(out)
}

// remote returns the remotes entry by key (gjson paths treat @ specially)
func remote(body gjson.Result, key string) gjson.Result {
	return body.Get("config.remotes").Map()[key]
}

func assertPreserved(t *testing.T, body gjson.Result) {
	assert.JSONEq(t, `{"devicemanagement": ["LWM2Module"]}`, body.Get("config.excludedRemotes").Raw)
	assert.JSONEq(t, `{"nested":[1,2,3]}`, body.Get("config.unknownField").Raw)
	assert.Equal(t, "123", body.Get("id").String())
	assert.JSONEq(t, `["DeviceManagementModule"]`, remote(body, "devicemanagement").Raw)
}

func Test_InstallPreservesConfigAndUnversionedRemotes(t *testing.T) {
	app := gjson.Parse(testApp)
	remotes, preserved := existing(app)
	require.NoError(t, buildUIRemotes(remotes, gjson.Parse(testPlugin), "latest"))

	body := build(t, remotes, preserved)
	assertPreserved(t, body)
	assert.JSONEq(t, `["myModule"]`, remote(body, "my-plugin@2.0.0").Raw)
	assert.JSONEq(t, `["asm"]`, remote(body, "c8y-asm-ui@3.0.1").Raw)
	assert.JSONEq(t, `["otherModule"]`, remote(body, "other-plugin@1.0.0").Raw)
}

func Test_InstallNewVersionReplacesOnlyThatPlugin(t *testing.T) {
	app := gjson.Parse(`{"id":"123","config":{"remotes":{"my-plugin@1.0.0":["myModule"],"other-plugin@1.0.0":["otherModule"]}}}`)
	remotes, preserved := existing(app)
	require.NoError(t, buildUIRemotes(remotes, gjson.Parse(testPlugin), "2.0.0"))

	out, err := buildApplicationBody(app, remotes, preserved, []byte("{}"))
	require.NoError(t, err)
	body := gjson.ParseBytes(out)
	assert.False(t, remote(body, "my-plugin@1.0.0").Exists())
	assert.True(t, remote(body, "my-plugin@2.0.0").Exists())
	assert.JSONEq(t, `["otherModule"]`, remote(body, "other-plugin@1.0.0").Raw)
}

func Test_DeleteRemovesOnlyNamedPlugin(t *testing.T) {
	app := gjson.Parse(testApp)
	remotes, preserved := existing(app)
	delete(remotes, "other-plugin")

	body := build(t, remotes, preserved)
	assertPreserved(t, body)
	assert.False(t, remote(body, "other-plugin@1.0.0").Exists())
	assert.JSONEq(t, `["asm"]`, remote(body, "c8y-asm-ui@3.0.1").Raw)
}

func Test_ConfigWithoutRemotes(t *testing.T) {
	app := gjson.Parse(`{"id":"9","config":{"excludedRemotes":{"devicemanagement":["LWM2Module"]}}}`)
	remotes, preserved := existing(app)
	require.NoError(t, buildUIRemotes(remotes, gjson.Parse(testPlugin), "latest"))

	out, err := buildApplicationBody(app, remotes, preserved, []byte("{}"))
	require.NoError(t, err)
	body := gjson.ParseBytes(out)
	assert.JSONEq(t, `{"devicemanagement":["LWM2Module"]}`, body.Get("config.excludedRemotes").Raw)
	assert.JSONEq(t, `["myModule"]`, remote(body, "my-plugin@2.0.0").Raw)
}
