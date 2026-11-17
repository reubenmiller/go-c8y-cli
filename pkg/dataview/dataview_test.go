package dataview

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tidwall/gjson"
)

func writeViews(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "views.json"), []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestAutoMatchDisabledViewIsOnlyUsedByName(t *testing.T) {
	dir := writeViews(t, `{
		"version": "1",
		"definitions": [
			{"name": "manualOnly", "priority": 100, "autoMatch": false, "columns": ["id"]},
			{"name": "fallback", "priority": 200, "columns": ["name"]}
		]
	}`)

	view, err := NewDataView(".*", ".json", nil, dir)
	if err != nil {
		t.Fatal(err)
	}

	data := gjson.Parse(`{"id": "1", "name": "test"}`)
	columns, err := view.GetView(&ViewData{ResponseBody: &data})
	if err != nil {
		t.Fatal(err)
	}
	if len(columns) != 1 || columns[0] != "name" {
		t.Errorf("expected the auto matched view to be used. got=%v", columns)
	}

	columns, err = view.GetViewByName("manualOnly")
	if err != nil {
		t.Fatal(err)
	}
	if len(columns) != 1 || columns[0] != "id" {
		t.Errorf("expected the view to be found by name. got=%v", columns)
	}
}

func TestSetActiveViewTakesPrecedence(t *testing.T) {
	dir := writeViews(t, `{"version": "1", "definitions": [{"name": "any", "columns": ["name"]}]}`)
	view, err := NewDataView(".*", ".json", nil, dir)
	if err != nil {
		t.Fatal(err)
	}

	view.SetActiveView(&Definition{Name: "custom", Columns: []string{"status"}})
	data := gjson.Parse(`{"id": "1", "name": "test"}`)
	columns, err := view.GetView(&ViewData{ResponseBody: &data})
	if err != nil {
		t.Fatal(err)
	}
	if len(columns) != 1 || columns[0] != "status" {
		t.Errorf("expected the active view to be used. got=%v", columns)
	}
}

func TestGetDefinitionByNameWithNamespace(t *testing.T) {
	dir := writeViews(t, `{"version": "1", "definitions": [{"name": "health", "columns": ["status"]}]}`)
	view, err := NewDataView(".*", ".json", nil, "myext"+NamespaceSeparator+dir)
	if err != nil {
		t.Fatal(err)
	}

	definition, err := view.GetDefinitionByName("myext::health")
	if err != nil {
		t.Fatal(err)
	}
	if definition == nil || definition.Extension != "myext" {
		t.Fatalf("expected the view from the extension. got=%v", definition)
	}

	definition, err = view.GetDefinitionByName("other::health")
	if err != nil {
		t.Fatal(err)
	}
	if definition != nil {
		t.Errorf("expected no view. got=%v", definition)
	}
}
