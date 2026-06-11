package tableviewer_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/tableviewer"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/encode"
)

// TestStreamWriterRendersFromSDKPipeline drives the full split: the go-c8y/v2
// table engine resolves columns and samples widths, and tablewriter v1
// renders the rows in streaming mode.
func TestStreamWriterRendersFromSDKPipeline(t *testing.T) {
	body := `{"managedObjects": [
		{"id": "1", "name": "alpha", "c8y_Hardware": {"model": "RPi4"}},
		{"id": "2", "name": "beta-much-longer", "c8y_Hardware": {"model": "NUC"}},
		{"id": "3", "name": "gamma", "c8y_Hardware": {"model": "X1"}}
	]}`

	var buf bytes.Buffer
	err := output.Render(context.Background(),
		output.FromBytes([]byte(body), "managedObjects"),
		encode.NewTableWithWriter(
			tableviewer.NewStreamWriter(&buf, false),
			encode.TableOptions{
				Columns:    []string{"id", "name", "c8y_Hardware.model"},
				SampleSize: 2,
			}))
	if err != nil {
		t.Fatalf("render failed: %s", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected header + separator + 3 rows, got %d lines:\n%s", len(lines), out)
	}
	for _, want := range []string{"| id ", "| alpha", "| beta-much-longer", "| gamma", "RPi4"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if !strings.HasPrefix(lines[1], "|:") && !strings.HasPrefix(lines[1], "|-") {
		t.Errorf("second line must be the markdown separator, got %q", lines[1])
	}
	// All rows must share the same rendered width (streaming with fixed
	// widths), including the row that arrived after the sample window.
	for i := 1; i < len(lines); i++ {
		if len(lines[i]) != len(lines[0]) {
			t.Errorf("line %d width %d != header width %d:\n%s", i, len(lines[i]), len(lines[0]), out)
		}
	}
}
