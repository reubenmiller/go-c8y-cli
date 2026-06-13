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
// TODO: table output needs a renderer wrapping the existing console/dataview
// machinery (sampled column detection); until then it falls back to JSON.
func NewRenderer(out io.Writer, cfg *config.Config) (output.Renderer, error) {
	var r output.Renderer
	switch cfg.GetOutputFormat() {
	case config.OutputCSV:
		r = encode.NewCSV(out, encode.CSVOptions{})
	case config.OutputCSVWithHeader:
		r = encode.NewCSV(out, encode.CSVOptions{Header: true})
	case config.OutputTSV:
		r = encode.NewTSV(out, encode.CSVOptions{})
	default:
		// json, serverresponse and (for now) table
		r = encode.NewNDJSON(out)
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
