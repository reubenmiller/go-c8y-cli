package c8ystream

import (
	"io"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonfilter"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/tidwall/gjson"
)

// consoleRenderer adapts the v1 pkg/console writer to the v2 output.Renderer
// interface. It reuses the existing console/dataview machinery so the
// terminal rendering — table column sampling/width/number-formatting/wrapping
// and JSON pretty-print/compact/colorisation — is byte-for-byte identical to
// the v1 commands.
//
// The client-side --select/--filter/--outputTemplate stages already ran in the
// pipeline, so each document reaching the renderer is the final shape. For
// table output the columns are the document's leaf paths: they are fed to the
// console as one resolved key group per row, and the console merges them
// across the sampled rows (so a column present only in a later row is still
// shown) exactly as the v1 output path does.
type consoleRenderer struct {
	console *console.Console
	isTable bool
}

// newConsoleRenderer builds a console renderer for the json/table/server
// response formats from the session/flag configuration.
func newConsoleRenderer(out io.Writer, cfg *config.Config) *consoleRenderer {
	tableOptions := &console.TableOptions{
		MinColumnWidth:           cfg.ViewColumnMinWidth(),
		MinEmptyValueColumnWidth: cfg.ViewColumnEmptyValueMinWidth(),
		MaxColumnWidth:           cfg.ViewColumnMaxWidth(),
		ColumnPadding:            cfg.ViewColumnPadding(),
		RowMode:                  cfg.ViewRowMode(),
		NumberFormatter:          cfg.GetTableViewNumberFormatter(),
		SampleSize:               cfg.ViewSampleSize(),
		SampleTimeout:            cfg.ViewSampleTimeout(),
	}
	c := console.NewConsole(out, tableOptions, nil)
	c.Format = cfg.GetOutputFormat()
	c.Compact = cfg.CompactJSON()
	c.Colorized = !cfg.DisableColor()
	return &consoleRenderer{
		console: c,
		isTable: c.IsTable(),
	}
}

func (r *consoleRenderer) Write(doc jsondoc.JSONDoc) error {
	raw := doc.Raw()
	if r.isTable {
		// Provide the resolved columns for this row so the console can merge
		// them into the table header across the sampled rows.
		if cols := leafPaths(raw); len(cols) > 0 {
			r.console.SetHeaderFromInput("", []jsonfilter.KeyGroup{{Pattern: "*", Keys: cols}})
		}
	}
	_, err := r.console.Write(raw)
	return err
}

func (r *consoleRenderer) Close() error {
	// Render any rows buffered while the table columns were being sampled.
	r.console.Flush()
	return nil
}

// leafPaths returns the dotted key paths of every leaf in a JSON object, in
// document order. Objects are traversed; arrays and scalars are leaves —
// matching how the select stage and the table viewer treat the document, so
// the returned paths can be looked up directly as table columns.
func leafPaths(raw []byte) []string {
	var paths []string
	var walk func(res gjson.Result, prefix string)
	walk = func(res gjson.Result, prefix string) {
		res.ForEach(func(k, v gjson.Result) bool {
			path := k.String()
			if prefix != "" {
				path = prefix + "." + path
			}
			if v.IsObject() {
				walk(v, path)
			} else {
				paths = append(paths, path)
			}
			return true
		})
	}
	walk(gjson.ParseBytes(raw), "")
	return paths
}
