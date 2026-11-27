package c8ystream

import (
	"io"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/encode"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/shape"
)

// shapingRenderer shapes each document with the columns of the output (see
// columnSelector) before delegating to the wrapped renderer. It lets the
// --outputFile JSON tee receive the shaped (selected) document even when the
// main stdout path skips the selectStage (table/CSV/TSV + --select or a view),
// because the renderer owns the columns there.
type shapingRenderer struct {
	output.Renderer
	cols    *columnSelector
	flatten bool
}

func (s *shapingRenderer) Write(doc jsondoc.JSONDoc) error {
	sel := s.cols.get()
	if sel == nil {
		return s.Renderer.Write(doc)
	}
	selection, err := sel.Apply(doc.Raw())
	if err != nil {
		return s.Renderer.Write(doc)
	}
	var b []byte
	if s.flatten {
		b, err = selection.FlatJSON()
	} else {
		b, err = selection.JSON()
	}
	if err != nil {
		return s.Renderer.Write(doc)
	}
	return s.Renderer.Write(jsondoc.New(b))
}

// columnsInRenderer reports whether the renderer owns the columns of the
// output, instead of a shaping stage dropping the properties a document lacks:
// the tabular formats, where --select and the view columns are columns (an
// empty cell for a document without the property), as in v1. The json formats
// shape the documents in a stage.
func columnsInRenderer(cfg *config.Config) bool {
	switch cfg.GetOutputFormat() {
	case config.OutputCSV, config.OutputCSVWithHeader, config.OutputTSV, config.OutputTable:
		// the columns are the same with --flatten (the renderer flattens), see
		// flattenUnlessColumns
		return true
	default:
		return false
	}
}

// newColumnSelector returns the selector of the columns of tabular output,
// resolved from --select; the view fills it in once it is known.
func newColumnSelector(cfg *config.Config) *columnSelector {
	cols := &columnSelector{}
	if selectors := cfg.GetJSONSelect(); len(selectors) > 0 {
		cols.set(selectors)
	}
	return cols
}

// selectionCSVActive reports whether the output should be rendered through the
// selection-aware CSV path: a CSV/TSV format together with an explicit
// --select or a view (which can apply). In that combination the selector, not
// the generic leaf-path CSV encoder, owns the columns, so array elements
// flatten to indexed columns (body.measurements.0.time) and an unmatched
// column keeps a stable empty column, matching the v1 jsonfilter path.
func selectionCSVActive(cfg *config.Config) bool {
	switch cfg.GetOutputFormat() {
	case config.OutputCSV, config.OutputCSVWithHeader, config.OutputTSV:
	default:
		return false
	}
	if len(cfg.GetJSONSelect()) > 0 {
		return true
	}
	return !cfg.RawOutput() && !strings.EqualFold(cfg.ViewOption(), config.ViewsOff)
}

// selectionCSVRenderer renders each document by applying the --select selector
// and writing the resulting Selection as a delimiter-separated row, in
// selection order. Unlike encode.CSV it derives the columns from the selection
// (Selection.Keys/Selection.CSV) rather than from the leaf paths of the shaped
// document, so:
//   - an array selected with ** yields one column per element
//     (e.g. body.measurements.0.time), instead of the whole array JSON in a
//     single column, and
//   - a --select pattern that matches nothing still produces a stable empty
//     column (the trailing-empty-column case).
//
// It works on the raw document, so the generic selectStage is skipped for this
// output (see CommonStages); aliases (alias:path) therefore resolve once, here.
type selectionCSVRenderer struct {
	w          io.Writer
	cols       *columnSelector
	fallback   output.Renderer // the generic encoder, while no columns are resolved
	separator  string
	header     bool
	headerDone bool
}

func newSelectionCSVRenderer(w io.Writer, cfg *config.Config, cols *columnSelector) *selectionCSVRenderer {
	separator := ","
	header := false
	switch cfg.GetOutputFormat() {
	case config.OutputTSV:
		separator = "\t"
	case config.OutputCSVWithHeader:
		header = true
	}
	var fallback output.Renderer
	if cfg.GetOutputFormat() == config.OutputTSV {
		fallback = encode.NewTSV(w, encode.CSVOptions{AutoFlush: true})
	} else {
		fallback = encode.NewCSV(w, encode.CSVOptions{Header: header, AutoFlush: true})
	}
	return &selectionCSVRenderer{
		w:         w,
		cols:      cols,
		fallback:  fallback,
		separator: separator,
		header:    header,
	}
}

func (e *selectionCSVRenderer) Write(doc jsondoc.JSONDoc) error {
	sel := e.cols.get()
	if sel == nil {
		// no view applies (yet): every property of the document
		return e.fallback.Write(doc)
	}
	selection, err := sel.Apply(doc.Raw())
	if err != nil {
		// A document the selector cannot parse passes through unchanged, as in v1.
		_, werr := e.w.Write(append(doc.Raw(), '\n'))
		return werr
	}
	if e.header && !e.headerDone {
		e.headerDone = true
		if _, err := io.WriteString(e.w, e.headerRow(selection)+"\n"); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(e.w, selection.CSV(e.separator)+"\n"); err != nil {
		return err
	}
	return nil
}

func (e *selectionCSVRenderer) Close() error { return e.fallback.Close() }

// headerRow joins the selection's display keys (alias applied, the integer-key
// marker stripped) with the separator.
func (e *selectionCSVRenderer) headerRow(selection *shape.Selection) string {
	keys := selection.Keys()
	display := make([]string, len(keys))
	for i, k := range keys {
		display[i] = strings.ReplaceAll(k, "::k::", "")
	}
	return strings.Join(display, e.separator)
}
