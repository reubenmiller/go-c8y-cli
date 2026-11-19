package c8ystream

import (
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
)

// CommonStages builds the standard client-side processing chain from the
// session/flag configuration: --filter → --outputTemplate → --select.
// Stages are only created for options the user actually set: an empty
// template is not a valid jsonnet snippet, and an empty selector would strip
// every property.
func CommonStages(cfg *config.Config, outputTemplate output.Stage) ([]output.Stage, error) {
	stages := make([]output.Stage, 0, 3)

	if expressions := cfg.GetJSONFilter(); len(expressions) > 0 {
		pred, err := filter.Parse(expressions...)
		if err != nil {
			return nil, err
		}
		stages = append(stages, output.Filter(pred))
	}

	// --outputTemplate runs between --filter and --select. The stage is built by
	// the runner (Runner.outputTemplateStage) so it can bind the per-item input,
	// command flags and request/response metadata; nil when no template is set.
	if outputTemplate != nil {
		stages = append(stages, outputTemplate)
	}

	if selectors := cfg.GetJSONSelect(); len(selectors) > 0 {
		// CSV/TSV + --select is rendered by the selection-aware CSV renderer
		// (NewRenderer), which owns the columns directly from the selector; adding
		// the generic selectStage here would shape the document first and drop the
		// array indices / unmatched columns the renderer needs.
		if !selectCSVActive(cfg) {
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
// to --outputFile as a JSON array when set.
//
// csv/tsv use the v2 encoders (columns derived from the documents, composing
// with --select); json/table/serverresponse delegate to the v1 console writer
// (consoleRenderer) so terminal rendering — table column detection and JSON
// pretty-print/compact/colorisation — matches the v1 commands exactly.
func NewRenderer(out io.Writer, cfg *config.Config) (output.Renderer, error) {
	var r output.Renderer
	switch {
	// CSV/TSV + --select: the selection owns the columns (array indices become
	// columns, an unmatched --select keeps a stable empty column), so bypass the
	// generic leaf-path CSV encoder. The generic selectStage is skipped for this
	// case (see CommonStages); this renderer applies the selector itself.
	case selectCSVActive(cfg):
		r = newSelectionCSVRenderer(out, cfg)
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
		r = newConsoleRenderer(out, cfg)
	}

	if path := cfg.GetOutputFile(); path != "" {
		var fileRenderer output.Renderer
		if strings.Contains(path, "{") && strings.Contains(path, "}") {
			// {id}/{name}/{type}/{owner} placeholders are filled from the first
			// document (v1 expanded them from the response body), so file creation
			// must wait until a document arrives.
			fileRenderer = &lazyFileRenderer{path: path}
		} else {
			f, err := createOutputFile(path)
			if err != nil {
				return nil, err
			}
			fileRenderer = &closingRenderer{Renderer: encode.NewJSONArray(f), closer: f}
		}
		// CSV/TSV + --select skips the generic selectStage, so the document
		// reaching the renderer is unshaped. The --outputFile tee writes JSON, so
		// shape it here (the same nested selection the json sink would have
		// received) to keep the file output consistent with --select.
		if selectCSVActive(cfg) {
			fileRenderer = &shapingRenderer{Renderer: fileRenderer, stage: selectStage(cfg, cfg.GetJSONSelect())}
		}
		r = Multi(r, fileRenderer)
	}

	return r, nil
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

// lazyFileRenderer defers the --outputFile creation until the first document
// so {id}/{name}/{type}/{owner} placeholders in the path can be filled from
// that document — the same fields (and value sanitisation) v1 extracted from
// the response body. With no documents no file is created.
type lazyFileRenderer struct {
	path string
	r    output.Renderer
	f    *os.File
}

func (l *lazyFileRenderer) Write(doc jsondoc.JSONDoc) error {
	if l.r == nil {
		path := l.path
		for _, field := range []string{"id", "name", "type", "owner"} {
			placeholder := "{" + field + "}"
			if strings.Contains(path, placeholder) {
				path = strings.ReplaceAll(path, placeholder, fileutilities.SanitizeFileName(doc.Get(field).String()))
			}
		}
		f, err := createOutputFile(path)
		if err != nil {
			return err
		}
		l.f = f
		l.r = encode.NewJSONArray(f)
	}
	return l.r.Write(doc)
}

func (l *lazyFileRenderer) Close() error {
	if l.r == nil {
		return nil
	}
	err := l.r.Close()
	if cerr := l.f.Close(); err == nil {
		err = cerr
	}
	return err
}

// closingRenderer closes an underlying resource (e.g. the --outputFile
// handle) after the renderer has flushed its trailing content.
type closingRenderer struct {
	output.Renderer
	closer io.Closer
}

func (c *closingRenderer) Close() error {
	err := c.Renderer.Close()
	if cerr := c.closer.Close(); err == nil {
		err = cerr
	}
	return err
}
