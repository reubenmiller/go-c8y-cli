package tableviewer_test

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/tableviewer"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/encode"
)

func TestCellFormatterArrays(t *testing.T) {
	doc := jsondoc.New([]byte(`{
		"tags": ["pump", "site-a"],
		"thresholds": [10, 20.5],
		"flags": [true, false, null],
		"empty": [],
		"objects": [{"id": "1"}],
		"nested": [[1, 2]],
		"text": "one, two"
	}`))
	format := tableviewer.CellFormatter(nil)

	cases := map[string]string{
		"tags":       "[pump, site-a]",
		"thresholds": "[10, 20.5]",
		"flags":      "[true, false, null]",
		"empty":      "[]",
		"objects":    `[{"id": "1"}]`,
		"nested":     "[[1, 2]]",
		"text":       "one, two",
		"missing":    "",
	}
	for column, want := range cases {
		got, align := format(doc, column)
		if got != want {
			t.Errorf("column %q: got %q, want %q", column, got, want)
		}
		if align != encode.AlignLeft {
			t.Errorf("column %q: expected left alignment", column)
		}
	}
}
