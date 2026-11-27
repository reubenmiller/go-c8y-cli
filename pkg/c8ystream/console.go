package c8ystream

import (
	"bytes"
	"io"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonfilter"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/shape"
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

	// selected are the --select patterns; an element of an array which was
	// selected explicitly (e.g. tags.1) gets its own table column
	selected []string

	// cols owns the columns of a table (--select, or the view): each document
	// is shaped by it here, and every column of it is a table column, also for
	// documents which lack it. Nil, or without a selector, the columns are the
	// properties of the documents
	cols *columnSelector
}

// newConsoleRenderer builds a console renderer for the json/table/server
// response formats from the session/flag configuration.
func newConsoleRenderer(out io.Writer, cfg *config.Config, cols *columnSelector) *consoleRenderer {
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
		console:  c,
		isTable:  c.IsTable(),
		selected: cfg.GetJSONSelect(),
		cols:     cols,
	}
}

func (r *consoleRenderer) Write(doc jsondoc.JSONDoc) error {
	raw := doc.Raw()
	if selection := r.cols.selection(); r.isTable && selection != nil {
		if shaped, groups, ok := shapeToColumns(selection, raw); ok {
			raw = shaped
			r.console.SetHeaderFromInput("", groups)
		}
	} else if r.isTable {
		// Provide the resolved columns for this row so the console can merge
		// them into the table header across the sampled rows.
		if cols := leafPaths(raw, r.selected); len(cols) > 0 {
			r.console.SetHeaderFromInput("", []jsonfilter.KeyGroup{{Pattern: "*", Keys: cols}})
		}
	}
	// console.Write only newline-terminates JSON objects/arrays; a scalar or
	// string document (e.g. a template producing `63613,agent01` or a stand-alone
	// timestamp) is written verbatim, so successive scalar docs would run
	// together. Terminate it here, matching v1's jsonformatter text path.
	if trimmed := bytes.TrimSpace(raw); !r.isTable && len(trimmed) > 0 &&
		!jsonUtilities.IsJSONObject(trimmed) && !jsonUtilities.IsJSONArray(trimmed) {
		raw = append(bytes.Clone(raw), '\n')
	}
	_, err := r.console.Write(raw)
	return err
}

// shapeToColumns applies the columns of a table (--select or a view) to a
// document. It returns the shaped document and the columns, grouped by the
// pattern which resolved them in selection order, so the table keeps a column
// for a pattern which the document lacks (shown as an empty cell, or as the
// pattern itself when no sampled row has it), as v1 did.
func shapeToColumns(columns *columnSelection, raw []byte) ([]byte, []jsonfilter.KeyGroup, bool) {
	selection, err := columns.selector.Apply(raw)
	if err != nil {
		return nil, nil, false
	}
	shaped, err := selection.JSON()
	if err != nil {
		return nil, nil, false
	}
	return shaped, assignColumns(leafPaths(shaped, columns.patterns), selection.Groups()), true
}

// assignColumns puts each column of a shaped document in the group of the first
// pattern which resolved it, keeping the order of the patterns. A pattern which
// resolved nothing keeps an empty group (the merged header names the pattern).
func assignColumns(paths []string, groups []shape.KeyGroup) []jsonfilter.KeyGroup {
	result := make([]jsonfilter.KeyGroup, len(groups))
	normalised := make([][]string, len(groups))
	for i, group := range groups {
		result[i].Pattern = group.Pattern
		for _, key := range group.Keys {
			normalised[i] = append(normalised[i], strings.ToLower(strings.ReplaceAll(key, "::k::", "")))
		}
	}
	for _, path := range paths {
		lower := strings.ToLower(path)
		assigned := false
		for i := range groups {
			for _, key := range normalised[i] {
				// the column is the key, below the key (an object), or the whole
				// array of which the key is an element
				if lower == key || strings.HasPrefix(lower, key+".") || strings.HasPrefix(key, lower+".") {
					result[i].Keys = append(result[i].Keys, path)
					assigned = true
					break
				}
			}
			if assigned {
				break
			}
		}
		if !assigned {
			result = append(result, jsonfilter.KeyGroup{Pattern: path, Keys: []string{path}})
		}
	}
	return result
}

func (r *consoleRenderer) Close() error {
	// Render any rows buffered while the table columns were being sampled.
	r.console.Flush()
	return nil
}

// leafPaths returns the dotted key paths of every column of a JSON object, in
// document order. Objects are traversed. Arrays of simple values (strings,
// numbers, booleans, null) are one column, shown as a bracketed list. Arrays
// which contain objects or arrays have a column per index (items.0.id), and an
// array element which --select named explicitly (tags.1) is its own column.
func leafPaths(raw []byte, selected []string) []string {
	var paths []string
	var walk func(res gjson.Result, prefix string)
	walk = func(res gjson.Result, prefix string) {
		res.ForEach(func(k, v gjson.Result) bool {
			path := k.String()
			if prefix != "" {
				path = prefix + "." + path
			}
			switch {
			case v.IsObject():
				if len(v.Map()) == 0 {
					// an empty object (c8y_IsDevice: {}) is a value, shown as {}
					paths = append(paths, path)
				} else {
					walk(v, path)
				}
			case v.IsArray():
				if elements := selectedElements(path, selected, v); len(elements) > 0 {
					paths = append(paths, elements...)
				} else if hasNestedValues(v) {
					walk(v, path)
				} else {
					paths = append(paths, path)
				}
			default:
				paths = append(paths, path)
			}
			return true
		})
	}
	walk(gjson.ParseBytes(raw), "")
	return paths
}

// hasNestedValues reports whether an array contains objects or arrays.
func hasNestedValues(arr gjson.Result) bool {
	nested := false
	arr.ForEach(func(_, v gjson.Result) bool {
		nested = v.IsObject() || v.IsArray()
		return !nested
	})
	return nested
}

// selectedElements returns the --select patterns which address an element of
// the array at path (e.g. tags.1 for tags) and which exist in the array.
func selectedElements(path string, selected []string, arr gjson.Result) []string {
	var elements []string
	prefix := path + "."
	for _, sel := range selected {
		rest, ok := strings.CutPrefix(sel, prefix)
		if !ok || rest == "" || rest[0] < '0' || rest[0] > '9' {
			continue
		}
		if arr.Get(rest).Exists() {
			elements = append(elements, sel)
		}
	}
	return elements
}
