package dataview

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestColumnsUnmarshalJSON(t *testing.T) {
	var cols Columns
	err := json.Unmarshal([]byte(`[
		"id",
		"model:c8y_Hardware.model",
		{"path": "description", "width": 20},
		{"name": "status", "path": "c8y_Availability.status", "minWidth": 5, "maxWidth": 10}
	]`), &cols)
	require.NoError(t, err)

	assert.Equal(t, Columns{
		{Path: "id"},
		{Path: "model:c8y_Hardware.model"},
		{Path: "description", Width: 20},
		{Name: "status", Path: "c8y_Availability.status", MinWidth: 5, MaxWidth: 10},
	}, cols)
	assert.Equal(t, []string{"id", "model:c8y_Hardware.model", "description", "status:c8y_Availability.status"}, cols.Selectors())
	assert.True(t, cols.HasWidthRules())
}

func TestColumnsUnmarshalJSONInvalid(t *testing.T) {
	cases := map[string]string{
		"missing path":           `[{"width": 10}]`,
		"unknown property":       `[{"path": "id", "widht": 10}]`,
		"negative width":         `[{"path": "id", "width": -1}]`,
		"min greater than max":   `[{"path": "id", "minWidth": 10, "maxWidth": 5}]`,
		"width on exclusion":     `[{"path": "!id", "width": 10}]`,
		"not a string or object": `[10]`,
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			var cols Columns
			assert.Error(t, json.Unmarshal([]byte(input), &cols))
		})
	}
}

func TestColumnsMarshalJSONRoundTrip(t *testing.T) {
	cols := Columns{{Path: "id"}, {Path: "description", Width: 20}}
	out, err := json.Marshal(cols)
	require.NoError(t, err)
	assert.JSONEq(t, `["id", {"path": "description", "width": 20}]`, string(out))
}

func TestColumnsUnmarshalYAML(t *testing.T) {
	var cols Columns
	err := yaml.Unmarshal([]byte(`
- id
- liveness state:components.livenessState
- path: description
  width: 20
- name: model
  path: c8y_Hardware.model
  maxWidth: 12
`), &cols)
	require.NoError(t, err)
	assert.Equal(t, Columns{
		{Path: "id"},
		{Path: "liveness state:components.livenessState"},
		{Path: "description", Width: 20},
		{Name: "model", Path: "c8y_Hardware.model", MaxWidth: 12},
	}, cols)

	var invalid Columns
	assert.Error(t, yaml.Unmarshal([]byte("- path: id\n  widht: 3\n"), &invalid))
	assert.Error(t, yaml.Unmarshal([]byte("- width: 3\n"), &invalid))
}

func TestColumnSpecMatches(t *testing.T) {
	cases := []struct {
		spec   ColumnSpec
		column string
		want   bool
	}{
		{ColumnSpec{Path: "id"}, "id", true},
		{ColumnSpec{Path: "id"}, "ID", true},
		{ColumnSpec{Path: "id"}, "name", false},
		{ColumnSpec{Path: "c8y_Hardware.model"}, "c8y_Hardware.model", true},
		{ColumnSpec{Name: "model", Path: "c8y_Hardware.model"}, "model", true},
		{ColumnSpec{Name: "model", Path: "c8y_Hardware.model"}, "c8y_Hardware.model", true},
		{ColumnSpec{Path: "c8y_Position.*"}, "c8y_Position.lat", true},
		{ColumnSpec{Path: "tags.**"}, "tags", true},
		{ColumnSpec{Path: "tags.**"}, "tags.0.name", true},
		{ColumnSpec{Path: "tags.**"}, "tagsOther", false},
		{ColumnSpec{Name: "args", Path: "arguments.**"}, "args.0", true},
		{ColumnSpec{Name: "args", Path: "arguments.**"}, "argsOther", false},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, tc.spec.Matches(tc.column), "spec=%+v column=%s", tc.spec, tc.column)
	}
}

func TestColumnsWidthRuleFirstMatchWins(t *testing.T) {
	cols := Columns{
		{Path: "id"},
		{Path: "c8y_Position.lat", Width: 5},
		{Path: "c8y_Position.*", Width: 10},
	}
	rule, ok := cols.WidthRule("c8y_Position.lat")
	require.True(t, ok)
	assert.Equal(t, 5, rule.Width)

	rule, ok = cols.WidthRule("c8y_Position.lng")
	require.True(t, ok)
	assert.Equal(t, 10, rule.Width)

	_, ok = cols.WidthRule("id")
	assert.False(t, ok, "columns without widths have no rule")
}
