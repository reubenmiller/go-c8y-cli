package c8ystream

import (
	"io"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/shape"
)

// shapingRenderer applies a shaping stage to each document before delegating to
// the wrapped renderer. It lets the --outputFile JSON tee receive the shaped
// (selected) document even when the main stdout path skips the selectStage
// (CSV/TSV + --select).
type shapingRenderer struct {
	output.Renderer
	stage output.Stage
}

func (s *shapingRenderer) Write(doc jsondoc.JSONDoc) error {
	return s.Renderer.Write(applyStageToDoc(s.stage, doc))
}

// selectCSVActive reports whether the output should be rendered through the
// selection-aware CSV path: a CSV/TSV format together with an explicit
// --select. In that combination the selector — not the generic leaf-path CSV
// encoder — owns the columns, so array elements flatten to indexed columns
// (body.measurements.0.time) and an unmatched trailing --select keeps a stable
// empty column, matching the v1 jsonfilter path.
func selectCSVActive(cfg *config.Config) bool {
	if len(cfg.GetJSONSelect()) == 0 {
		return false
	}
	switch cfg.GetOutputFormat() {
	case config.OutputCSV, config.OutputCSVWithHeader, config.OutputTSV:
		return true
	default:
		return false
	}
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
	sel        *shape.Selector
	separator  string
	header     bool
	headerDone bool
}

func newSelectionCSVRenderer(w io.Writer, cfg *config.Config) *selectionCSVRenderer {
	separator := ","
	header := false
	switch cfg.GetOutputFormat() {
	case config.OutputTSV:
		separator = "\t"
	case config.OutputCSVWithHeader:
		header = true
	}
	return &selectionCSVRenderer{
		w:         w,
		sel:       shape.NewSelector(cfg.GetJSONSelect()...),
		separator: separator,
		header:    header,
	}
}

func (e *selectionCSVRenderer) Write(doc jsondoc.JSONDoc) error {
	selection, err := e.sel.Apply(doc.Raw())
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

func (e *selectionCSVRenderer) Close() error { return nil }

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
