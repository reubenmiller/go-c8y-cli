package c8ystream

import (
	"io"
	"os"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/encode"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/filter"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/shape"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/template"
)

// CommonStages builds the standard client-side processing chain from the
// session/flag configuration: --filter → --outputTemplate → --select.
// Stages are only created for options the user actually set: an empty
// template is not a valid jsonnet snippet, and an empty selector would strip
// every property.
func CommonStages(cfg *config.Config) ([]output.Stage, error) {
	stages := make([]output.Stage, 0, 3)

	if expressions := cfg.GetJSONFilter(); len(expressions) > 0 {
		pred, err := filter.Parse(expressions...)
		if err != nil {
			return nil, err
		}
		stages = append(stages, output.Filter(pred))
	}

	if tpl := cfg.GetOutputTemplate(); tpl != "" {
		stage, err := template.Jsonnet(tpl)
		if err != nil {
			return nil, err
		}
		stages = append(stages, stage)
	}

	if selectors := cfg.GetJSONSelect(); len(selectors) > 0 {
		stages = append(stages, shape.Select(selectors...))
	}

	return stages, nil
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
	switch cfg.GetOutputFormat() {
	case config.OutputCSV:
		// AutoFlush: terminal output is streamed, so each row must appear as it
		// is produced (items may arrive seconds apart) rather than being held in
		// csv.Writer's buffer until Close — matching the json/table sinks.
		r = encode.NewCSV(out, encode.CSVOptions{AutoFlush: true})
	case config.OutputCSVWithHeader:
		r = encode.NewCSV(out, encode.CSVOptions{Header: true, AutoFlush: true})
	case config.OutputTSV:
		r = encode.NewTSV(out, encode.CSVOptions{AutoFlush: true})
	default:
		// json, serverresponse and table
		r = newConsoleRenderer(out, cfg)
	}

	if path := cfg.GetOutputFile(); path != "" {
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		r = Multi(r, &closingRenderer{Renderer: encode.NewJSONArray(f), closer: f})
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
