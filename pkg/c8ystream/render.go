package c8ystream

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileutilities"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/encode"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/filter"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/shape"
	"github.com/tidwall/gjson"
)

// CommonStages builds the standard client-side processing chain from the
// session/flag configuration: --filter → --outputTemplate → --select.
// Stages are only created for options the user actually set: an empty
// template is not a valid jsonnet snippet, and an empty selector would strip
// every property.
func CommonStages(cfg *config.Config, outputTemplate output.Stage) ([]output.Stage, error) {
	return commonStages(cfg, outputTemplate)
}

func commonStages(cfg *config.Config, outputTemplate output.Stage) ([]output.Stage, error) {
	stages := make([]output.Stage, 0, 3)

	if expressions := cfg.GetJSONFilter(); len(expressions) > 0 {
		pred, err := filter.Parse(expressions...)
		if err != nil {
			return nil, err
		}
		stages = append(stages, filterStage(cfg, pred))
	}

	// --outputTemplate runs between --filter and --select. The stage is built by
	// the runner (Runner.outputTemplateStage) so it can bind the per-item input,
	// command flags and request/response metadata; nil when no template is set.
	if outputTemplate != nil {
		stages = append(stages, outputTemplate)
	}

	if selectors := cfg.GetJSONSelect(); len(selectors) > 0 {
		// Table/CSV/TSV + --select is rendered by a selection-aware renderer
		// (NewRenderer), which owns the columns directly from the selector; adding
		// the generic selectStage here would shape the document first and drop the
		// array indices / unmatched columns the renderer needs.
		if !columnsInRenderer(cfg) {
			stages = append(stages, selectStage(cfg, selectors))
		}
	} else if cfg.FlattenJSON() {
		// --flatten without an explicit --select flattens the whole document
		// (v1 ran flatten through the same selector path with an implicit "**").
		// With a selector present the flattening happens inside selectStage.
		stages = append(stages, selectStage(cfg, []string{"**"}))
	}

	return stages, nil
}

// selectStage shapes each document with the full --select selector — aliases
// (alias:path), the ** globstar, array traversal and !negation — instead of the
// limited shape.Select stage, matching the v1 jsonfilter path. --flatten renders
// the selection as a flat dotted-key object; otherwise a nested document with
// keys in selection order. A document the selector cannot parse passes through
// unchanged, as in v1.
func selectStage(cfg *config.Config, selectors []string) output.Stage {
	sel := shape.NewSelector(selectors...)
	flatten := cfg.FlattenJSON()
	return output.Map(func(doc jsondoc.JSONDoc) (jsondoc.JSONDoc, error) {
		selection, err := sel.Apply(doc.Raw())
		if err != nil {
			return doc, nil
		}
		var b []byte
		if flatten {
			b, err = selection.FlatJSON()
		} else {
			b, err = selection.JSON()
		}
		if err != nil {
			return doc, nil
		}
		return jsondoc.New(b), nil
	})
}

// NewRenderer builds the sink for a command from the --output format, teed
// to --outputFile when set (see newFileTee for what the file holds).
//
// csv/tsv use the v2 encoders (columns derived from the documents, composing
// with --select); json/table/serverresponse delegate to the v1 console writer
// (consoleRenderer) so terminal rendering — table column detection and JSON
// pretty-print/compact/colorisation — matches the v1 commands exactly.
func NewRenderer(out io.Writer, cfg *config.Config) (output.Renderer, error) {
	return newRenderer(out, cfg, newColumnSelector(cfg), nil)
}

// newRenderer builds the sink for the output format. cols owns the columns of
// tabular output (--select, or the view once it is resolved), see columnSelector.
// received tells whether a response with a body was received (see newFileTee).
func newRenderer(out io.Writer, cfg *config.Config, cols *columnSelector, received func() bool) (output.Renderer, error) {
	var r output.Renderer
	switch {
	// CSV/TSV + --select: the selection owns the columns (array indices become
	// columns, an unmatched --select keeps a stable empty column), so bypass the
	// generic leaf-path CSV encoder. The generic selectStage is skipped for this
	// case (see CommonStages); this renderer applies the selector itself.
	case selectionCSVActive(cfg):
		r = newSelectionCSVRenderer(out, cfg, cols)
	case cfg.GetOutputFormat() == config.OutputCSV:
		// AutoFlush: terminal output is streamed, so each row must appear as it
		// is produced (items may arrive seconds apart) rather than being held in
		// csv.Writer's buffer until Close — matching the json/table sinks.
		r = encode.NewCSV(out, encode.CSVOptions{AutoFlush: true})
	case cfg.GetOutputFormat() == config.OutputCSVWithHeader:
		r = encode.NewCSV(out, encode.CSVOptions{Header: true, AutoFlush: true})
	case cfg.GetOutputFormat() == config.OutputTSV:
		r = encode.NewTSV(out, encode.CSVOptions{AutoFlush: true})
	default:
		// json, serverresponse and table
		r = newConsoleRenderer(out, cfg, cols)
	}

	if path := cfg.GetOutputFile(); path != "" && cfg.GetOutputFormat() != config.OutputServerResponse {
		r = Multi(r, newFileTee(path, cfg, cols, received))
	}

	return r, nil
}

// fileTee writes the documents to --outputFile as v2.59.0 did: what the
// command prints, without the header of csvheader, one document per line and
// no newline after the last one.
//
//   - csv, tsv and csvheader: the rows of the output;
//   - json and table: the documents as compact JSON (after --filter,
//     --outputTemplate and --select), also for table, which prints no JSON;
//   - serverresponse: no file.
//
// The file is created with the first document (which fills {id}, {name},
// {type} and {owner} in the path, one file for each distinct path) and
// replaced if it exists. A request which fails, a dry run and a response
// without a body write no file; a response with a body but no document (e.g.
// an empty list) writes an empty file.
type fileTee struct {
	path     string
	cfg      *config.Config
	cols     *columnSelector
	received func() bool
	sinks    map[string]*fileSink
	order    []string
}

type fileSink struct {
	f *os.File
	r output.Renderer
}

func newFileTee(path string, cfg *config.Config, cols *columnSelector, received func() bool) *fileTee {
	return &fileTee{path: path, cfg: cfg, cols: cols, received: received, sinks: map[string]*fileSink{}}
}

func (t *fileTee) hasPlaceholders() bool {
	return strings.Contains(t.path, "{") && strings.Contains(t.path, "}")
}

// pathFor fills {id}/{name}/{type}/{owner} from the document (v1 expanded
// them from the response body).
func (t *fileTee) pathFor(doc jsondoc.JSONDoc) string {
	path := t.path
	if !t.hasPlaceholders() {
		return path
	}
	for _, field := range []string{"id", "name", "type", "owner"} {
		placeholder := "{" + field + "}"
		if strings.Contains(path, placeholder) {
			path = strings.ReplaceAll(path, placeholder, fileutilities.SanitizeFileName(doc.Get(field).String()))
		}
	}
	return path
}

func (t *fileTee) sink(path string) (*fileSink, error) {
	if s, ok := t.sinks[path]; ok {
		return s, nil
	}
	f, err := createOutputFile(path)
	if err != nil {
		return nil, err
	}
	w := &noTrailingNewline{w: f}
	s := &fileSink{f: f}
	switch {
	case columnsInRenderer(t.cfg) && t.cfg.GetOutputFormat() != config.OutputTable:
		s.r = newTabularFileRenderer(w, t.cfg, t.cols)
	default:
		var r output.Renderer = &jsonLinesRenderer{w: w}
		// table + --select or a view: the document reaching the renderer is
		// unshaped, as the renderer owns the columns
		if columnsInRenderer(t.cfg) {
			r = &shapingRenderer{Renderer: r, cols: t.cols, flatten: t.cfg.FlattenJSON()}
		}
		s.r = r
	}
	t.sinks[path] = s
	t.order = append(t.order, path)
	return s, nil
}

func (t *fileTee) Write(doc jsondoc.JSONDoc) error {
	s, err := t.sink(t.pathFor(doc))
	if err != nil {
		return err
	}
	return s.r.Write(doc)
}

func (t *fileTee) Close() error {
	var firstErr error
	if len(t.sinks) == 0 && !t.hasPlaceholders() && t.received != nil && t.received() {
		// a response with a body but no document (an empty list)
		if _, err := t.sink(t.path); err != nil {
			firstErr = err
		}
	}
	for _, path := range t.order {
		s := t.sinks[path]
		if err := s.r.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		if err := s.f.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// newTabularFileRenderer is the csv, tsv or csvheader renderer of the file,
// which has no header row.
func newTabularFileRenderer(w io.Writer, cfg *config.Config, cols *columnSelector) output.Renderer {
	switch {
	case selectionCSVActive(cfg):
		r := newSelectionCSVRenderer(w, cfg, cols)
		r.header = false
		return r
	case cfg.GetOutputFormat() == config.OutputTSV:
		return encode.NewTSV(w, encode.CSVOptions{AutoFlush: true})
	default:
		return encode.NewCSV(w, encode.CSVOptions{AutoFlush: true})
	}
}

// jsonLinesRenderer writes each document as compact JSON on its own line; a
// document which is not JSON (the text of an --outputTemplate) is written as is.
type jsonLinesRenderer struct {
	w io.Writer
}

func (j *jsonLinesRenderer) Write(doc jsondoc.JSONDoc) error {
	raw := doc.Raw()
	if gjson.ValidBytes(raw) {
		var buf bytes.Buffer
		if err := json.Compact(&buf, raw); err == nil {
			raw = buf.Bytes()
		}
		// "<", ">" and "&" are written as the characters, as on stdout
		raw = bytes.ReplaceAll(raw, []byte("\\u003c"), []byte("<"))
		raw = bytes.ReplaceAll(raw, []byte("\\u003e"), []byte(">"))
		raw = bytes.ReplaceAll(raw, []byte("\\u0026"), []byte("&"))
	}
	_, err := j.w.Write(append(append([]byte{}, raw...), '\n'))
	return err
}

func (j *jsonLinesRenderer) Close() error { return nil }

// noTrailingNewline drops the newline after the last write: the lines of a
// file are separated by a newline, and the last one has none.
type noTrailingNewline struct {
	w       io.Writer
	pending bool
}

func (n *noTrailingNewline) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if n.pending {
		if _, err := n.w.Write([]byte("\n")); err != nil {
			return 0, err
		}
		n.pending = false
	}
	if p[len(p)-1] == '\n' {
		n.pending = true
		if _, err := n.w.Write(p[:len(p)-1]); err != nil {
			return 0, err
		}
		return len(p), nil
	}
	return n.w.Write(p)
}

// Multi fans every document out to all renderers — teeing at the sink, which
// allows each destination to use a different encoding (e.g. NDJSON on the
// terminal, a JSON array in a file).
func Multi(renderers ...output.Renderer) output.Renderer {
	return multiRenderer(renderers)
}

type multiRenderer []output.Renderer

func (m multiRenderer) Write(doc jsondoc.JSONDoc) error {
	for _, r := range m {
		if err := r.Write(doc); err != nil {
			return err
		}
	}
	return nil
}

func (m multiRenderer) Close() error {
	var firstErr error
	for _, r := range m {
		if err := r.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// createOutputFile creates the --outputFile target, creating missing parent
// directories first (v1's WriteToFile always ran MkdirAll on the dirname).
func createOutputFile(path string) (*os.File, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, err
		}
	}
	return os.Create(path)
}
