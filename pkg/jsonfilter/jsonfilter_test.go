package jsonfilter

import (
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/assert"
)

func Test_MergeKeyGroups(t *testing.T) {
	// patterns which did not resolve against the first row should be
	// replaced by the keys resolved from later rows
	columns := MergeKeyGroups([][]KeyGroup{
		{
			{Pattern: "name", Keys: []string{"name"}},
			{Pattern: "c8y_firmware.versio*", Keys: nil},
		},
		{
			{Pattern: "name", Keys: []string{"name"}},
			{Pattern: "c8y_firmware.versio*", Keys: []string{"c8y_Firmware.version"}},
		},
	})
	assert.EqualMarshalJSON(t, columns, `["name","c8y_Firmware.version"]`)

	// keys are deduplicated case insensitively, and patterns which never
	// resolve are included as-is
	columns = MergeKeyGroups([][]KeyGroup{
		{
			{Pattern: "c8y_firmware.*", Keys: []string{"c8y_Firmware.version"}},
			{Pattern: "missing", Keys: nil},
		},
		{
			{Pattern: "c8y_firmware.*", Keys: []string{"c8y_firmware.VERSION", "c8y_Firmware.url"}},
			{Pattern: "missing", Keys: nil},
		},
	})
	assert.EqualMarshalJSON(t, columns, `["c8y_Firmware.version","c8y_Firmware.url","missing"]`)
}

func Test_CollapseScalarArrays(t *testing.T) {
	doc := `{"id":"1","tags":["a","b","c"],"items":[{"id":"x"},{"id":"y"}],"mixed":["a",{"k":"v"}],"empty":[]}`
	groups := []KeyGroup{
		{Pattern: "id", Keys: []string{"id"}},
		{Pattern: "tags", Keys: []string{"tags.0", "tags.1", "tags.2"}},
		{Pattern: "items", Keys: []string{"items.0.id", "items.1.id"}},
		{Pattern: "mixed", Keys: []string{"mixed.0", "mixed.1.k"}},
		{Pattern: "empty", Keys: []string{"empty"}},
		{Pattern: "tags.1", Keys: []string{"tags.1"}},
		{Pattern: "missing", Keys: nil},
	}
	got := CollapseScalarArrays(groups, doc)
	want := [][]string{
		{"id"},
		{"tags"},
		{"items.0.id", "items.1.id"},
		{"mixed.0", "mixed.1.k"},
		{"empty"},
		{"tags.1"},
		{},
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d groups, got %d", len(want), len(got))
	}
	for i, group := range got {
		if group.Pattern != groups[i].Pattern {
			t.Errorf("group %d: pattern changed to %q", i, group.Pattern)
		}
		if strings.Join(group.Keys, ",") != strings.Join(want[i], ",") {
			t.Errorf("group %q: got %v, want %v", group.Pattern, group.Keys, want[i])
		}
	}
}
