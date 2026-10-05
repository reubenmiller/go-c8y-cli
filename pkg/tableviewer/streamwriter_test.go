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

// TestStreamWriterRightAlignsNumericColumns checks that numeric columns are
// marked as right-aligned in the markdown separator (---:) and that every
// row, including those after the first, is right-aligned.
func TestStreamWriterRightAlignsNumericColumns(t *testing.T) {
	body := `{"items": [
		{"name": "a", "value": 1},
		{"name": "b", "value": 22},
		{"name": "c", "value": 333}
	]}`

	var buf bytes.Buffer
	err := output.Render(context.Background(),
		output.FromBytes([]byte(body), "items"),
		encode.NewTableWithWriter(
			tableviewer.NewStreamWriter(&buf, false),
			encode.TableOptions{
				Columns: []string{"name", "value"},
			}))
	if err != nil {
		t.Fatalf("render failed: %s", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected header + separator + 3 rows, got %d lines:\n%s", len(lines), out)
	}
	if !strings.HasPrefix(lines[1], "|:-") || !strings.HasSuffix(lines[1], "-:|") {
		t.Errorf("expected left-aligned name and right-aligned value separator, got %q", lines[1])
	}
	for i, want := range []string{"   1 |", "  22 |", " 333 |"} {
		if !strings.HasSuffix(lines[i+2], want) {
			t.Errorf("row %d not right-aligned, got %q:\n%s", i, lines[i+2], out)
		}
	}
}
