package c8ystream

import (
	"testing"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/shape"
	"github.com/tidwall/gjson"
)

// TestApplyStageToDocSelectsExistingColumns checks that running one document
// through a shape.Select stage keeps every selected column that exists and
// silently drops the ones that don't — the behaviour the auto-view path relies
// on, where a view lists more columns than a given object carries.
func TestApplyStageToDocSelectsExistingColumns(t *testing.T) {
	doc := jsondoc.New([]byte(`{"id":"10025","name":"dev","owner":"admin","lastUpdated":"2026-01-01T00:00:00Z","self":"http://x"}`))
	// View columns: two of these (type, c8y_Availability.status) are absent.
	cols := []string{"id", "name", "type", "owner", "lastUpdated", "c8y_Availability.status"}

	out := applyStageToDoc(shape.Select(cols...), doc)
	got := gjson.ParseBytes(out.Raw())

	for _, present := range []string{"id", "name", "owner", "lastUpdated"} {
		if !got.Get(present).Exists() {
			t.Errorf("expected column %q to be kept, output=%s", present, out.Raw())
		}
	}
	for _, absent := range []string{"type", "self", "c8y_Availability"} {
		if got.Get(absent).Exists() {
			t.Errorf("expected column %q to be absent, output=%s", absent, out.Raw())
		}
	}
}
