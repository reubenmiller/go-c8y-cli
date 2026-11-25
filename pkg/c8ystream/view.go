package c8ystream

import (
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/tableviewer"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/shape"
	"github.com/tidwall/gjson"
)

// viewStage restricts every document to a Cumulocity "view" — a named,
// fragment-detected column set — the way the v1 request handler did before
// rendering. It shapes the documents (so the columns apply to json/csv/table
// alike, matching v1), and returns nil — no stage, documents pass through whole
// — when a view does not apply:
//
//   - --select was given (explicit selection wins),
//   - --raw / --withTotalPages etc. force the full envelope,
//   - --view off, or
//   - --view <name> / the empty default matches no definition.
//
// With --view auto the columns depend on the data, so detection is deferred to
// the first document; a named view is resolved up front.
func (r *Runner) viewStage() output.Stage {
	cfg := r.Config
	if len(cfg.GetJSONSelect()) > 0 || cfg.RawOutput() {
		return nil
	}

	view := cfg.ViewOption()
	if strings.EqualFold(view, config.ViewsOff) {
		return nil
	}

	dv, err := r.Factory.DataView()
	if err != nil || dv == nil {
		return nil
	}

	if strings.EqualFold(view, config.ViewsAuto) {
		return autoViewStage(dv, r.setViewDefinition)
	}

	// Named view; the empty default resolves to no match here, leaving the
	// document whole (all columns), which is the v1 default.
	cols, err := dv.GetViewByName(view)
	if err != nil || len(cols) == 0 {
		return nil
	}
	if definition, err := dv.GetDefinitionByName(view); err == nil {
		r.setViewDefinition(definition)
	}
	return viewSelectStage(cols)
}

// setViewDefinition applies the column width rules of the view to the table
// output (the columns of a view can set width/minWidth/maxWidth).
func (r *Runner) setViewDefinition(definition *dataview.Definition) {
	r.viewDefinition = definition
	r.applyViewDefinition()
}

// applyViewDefinition hands the column width rules of the active view to the
// console renderer, as soon as both are known.
func (r *Runner) applyViewDefinition() {
	if r.viewDefinition == nil || r.consoleRenderer == nil {
		return
	}
	if rules := tableviewer.ColumnWidthRules(r.viewDefinition); rules != nil {
		r.consoleRenderer.console.SetColumnWidthRules(rules)
	}
}

// findConsoleRenderer returns the console renderer of a renderer (which can be
// wrapped to also write --outputFile), or nil.
func findConsoleRenderer(r output.Renderer) *consoleRenderer {
	switch v := r.(type) {
	case *consoleRenderer:
		return v
	case multiRenderer:
		for _, item := range v {
			if c := findConsoleRenderer(item); c != nil {
				return c
			}
		}
	}
	return nil
}

// viewSelectStage shapes the documents to the columns of a view with the same
// selector as --select, so the view columns can use aliases (name:path),
// wildcards (tags.**) and exclusions, like the --select flag. Columns which a
// document does not have are dropped.
func viewSelectStage(cols []string) output.Stage {
	sel := shape.NewSelector(cols...)
	return output.Map(func(doc jsondoc.JSONDoc) (jsondoc.JSONDoc, error) {
		selection, err := sel.Apply(doc.Raw())
		if err != nil {
			return doc, nil
		}
		b, err := selection.JSON()
		if err != nil {
			return doc, nil
		}
		return jsondoc.New(b), nil
	})
}

// autoViewStage detects the view from the first document (by its fragments) and
// applies the detected columns to every document, so a streamed collection is
// shaped consistently. If no definition matches, the documents pass through
// unchanged.
func autoViewStage(dv *dataview.DataView, onDefinition func(*dataview.Definition)) output.Stage {
	return func(src output.Seq) output.Seq {
		return func(yield func(jsondoc.JSONDoc, error) bool) {
			var selector output.Stage
			detected := false
			for doc, err := range src {
				if err != nil {
					if !yield(doc, err) {
						return
					}
					continue
				}
				if !detected {
					detected = true
					selector = detectViewSelector(dv, doc)
					if selector != nil {
						onDefinition(dv.GetActiveView())
					}
				}
				out := doc
				if selector != nil {
					out = applyStageToDoc(selector, doc)
				}
				if !yield(out, nil) {
					return
				}
			}
		}
	}
}

// detectViewSelector returns the shaping stage for the view matching a
// document, or nil when none matches.
func detectViewSelector(dv *dataview.DataView, doc jsondoc.JSONDoc) output.Stage {
	res := gjson.ParseBytes(doc.Raw())
	cols, err := dv.GetView(&dataview.ViewData{ResponseBody: &res})
	if err != nil || len(cols) == 0 {
		return nil
	}
	return viewSelectStage(cols)
}

// applyStageToDoc runs a single document through a stage, returning the shaped
// document (the original on error). Used to reuse a once-built selector stage
// across the documents of an auto-detected view.
func applyStageToDoc(stage output.Stage, doc jsondoc.JSONDoc) jsondoc.JSONDoc {
	out := doc
	single := output.Seq(func(yield func(jsondoc.JSONDoc, error) bool) { yield(doc, nil) })
	for d, e := range stage(single) {
		if e == nil {
			out = d
		}
	}
	return out
}
