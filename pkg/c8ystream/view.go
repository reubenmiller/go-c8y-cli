package c8ystream

import (
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/tableviewer"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/shape"
	"github.com/tidwall/gjson"
)

// viewStage restricts every document to a Cumulocity "view" — a named,
// fragment-detected column set — the way the v1 request handler did before
// rendering. It returns nil — no stage, documents pass through whole — when a
// view does not apply:
//
//   - --select was given (explicit selection wins),
//   - --raw / --withTotalPages etc. force the full envelope,
//   - --view off, or
//   - --view <name> / the empty default matches no definition.
//
// The columns of a view are a property of the output, not of a document: like
// v1 (where a view became the --select of the request) a table or csv/tsv shows
// every column of the view, as an empty cell for the documents which lack it.
// So for those formats the stage only resolves the view and hands its selector
// to the renderer through cols. For json the documents are shaped here (a
// property a document lacks is simply absent).
//
// With --view auto the columns depend on the data, so the view is detected from
// the first document (as v1 did, with the request and the content type of its
// response); a named view is resolved up front.
func (r *Runner) viewStage(cols *columnSelector) output.Stage {
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

	shapeDocs := !columnsInRenderer(cfg)
	if strings.EqualFold(view, config.ViewsAuto) {
		return autoViewStage(dv, r.lastViewResponse, cols, shapeDocs, r.setViewDefinition)
	}

	// Named view; the empty default resolves to no match here, leaving the
	// document whole (all columns), which is the v1 default.
	columns, err := dv.GetViewByName(view)
	if err != nil || len(columns) == 0 {
		// as v1: a named view which does not exist is reported, and every column
		// is shown
		if log, logErr := r.Factory.Logger(); logErr == nil && log != nil {
			if err != nil {
				log.Warnf("no matching view found. %s, name=%s", err, view)
			} else {
				log.Warnf("no matching view found. name=%s", view)
			}
		}
		return nil
	}
	if definition, err := dv.GetDefinitionByName(view); err == nil {
		r.setViewDefinition(definition)
	}
	cols.set(columns)
	if !shapeDocs {
		return nil
	}
	return viewSelectStage(cols.get())
}

// viewResponse is the request and response header of a response the server
// sent, which view detection matches (request path/method, content type).
type viewResponse struct {
	request     *http.Request
	contentType string
}

// viewDetectionActive reports whether the view of the output is detected from
// the responses (--view auto, with no --select and no --raw).
func (r *Runner) viewDetectionActive() bool {
	cfg := r.Config
	if len(cfg.GetJSONSelect()) > 0 || cfg.RawOutput() {
		return false
	}
	return strings.EqualFold(cfg.ViewOption(), config.ViewsAuto)
}

// installViewCapture records the request and content type of the latest
// response, which --view auto matches the view definitions against (they can
// match on requestPath, requestMethod and contentType, like in v1). Like the
// --outputTemplate capture it relies on the serial pipeline: the middleware
// fires while a job runs its service call, before its documents are rendered,
// so the latest response is the one of the document being rendered.
func (r *Runner) installViewCapture(client *apiv2.Client) {
	if client == nil || !r.viewDetectionActive() {
		return
	}
	client.Use(func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			resp, err := next.RoundTrip(req)
			if err == nil && resp != nil {
				r.viewMu.Lock()
				r.viewResp = &viewResponse{request: req, contentType: resp.Header.Get("Content-Type")}
				r.viewMu.Unlock()
			}
			return resp, err
		})
	})
}

// lastViewResponse returns the latest response which the server sent, or nil
// (e.g. for a dry run, or when no request was made).
func (r *Runner) lastViewResponse() *viewResponse {
	r.viewMu.Lock()
	defer r.viewMu.Unlock()
	return r.viewResp
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

// findFileTee returns the --outputFile writer of a renderer, if any.
func findFileTee(r output.Renderer) *fileTee {
	switch v := r.(type) {
	case *fileTee:
		return v
	case multiRenderer:
		for _, item := range v {
			if t := findFileTee(item); t != nil {
				return t
			}
		}
	}
	return nil
}

// viewSelectStage shapes the documents to the columns of a view with the same
// selector as --select, so the view columns can use aliases (name:path),
// wildcards (tags.**) and exclusions, like the --select flag. Columns which a
// document does not have are dropped from the document.
func viewSelectStage(sel *shape.Selector) output.Stage {
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

// autoViewStage detects the view from the first document (by its fragments,
// and the request and content type of the response it came from) and applies
// the detected columns to every document, so a streamed collection is shaped
// consistently. As in v1, where the view of the first response became the
// --select of the following responses (also of piped input and of further
// pages), the detection runs once: if no definition matches the first
// document, the documents pass through unchanged.
func autoViewStage(dv *dataview.DataView, response func() *viewResponse, cols *columnSelector, shapeDocs bool, onDefinition func(*dataview.Definition)) output.Stage {
	return func(src output.Seq) output.Seq {
		return func(yield func(jsondoc.JSONDoc, error) bool) {
			var shaper output.Stage
			matched := false
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
					if columns := detectViewColumns(dv, doc, response()); len(columns) > 0 {
						cols.set(columns)
						matched = true
						shaper = viewSelectStage(cols.get())
						onDefinition(dv.GetActiveView())
					}
				}
				out := doc
				if matched && shapeDocs {
					out = applyStageToDoc(shaper, doc)
				}
				if !yield(out, nil) {
					return
				}
			}
		}
	}
}

// detectViewColumns returns the columns of the view matching a document (and
// the response it came from), or nil when none matches.
func detectViewColumns(dv *dataview.DataView, doc jsondoc.JSONDoc, resp *viewResponse) []string {
	res := gjson.ParseBytes(doc.Raw())
	data := &dataview.ViewData{ResponseBody: &res}
	if resp != nil {
		data.ContentType = resp.contentType
		data.Request = resp.request
	}
	columns, err := dv.GetView(data)
	if err != nil || len(columns) == 0 {
		return nil
	}
	return columns
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

// columnSelector is the selector which owns the columns of tabular output
// (table, csv, tsv): --select, or the columns of the view once it is resolved.
// The renderer applies it to every document, so a column which the document
// lacks is still a column (an empty cell), as it was in v1.
type columnSelector struct {
	current atomic.Pointer[columnSelection]
}

type columnSelection struct {
	patterns []string
	selector *shape.Selector
}

func (c *columnSelector) set(patterns []string) {
	if c == nil {
		return
	}
	c.current.Store(&columnSelection{patterns: patterns, selector: shape.NewSelector(patterns...)})
}

// get returns the selector, or nil while no columns are resolved.
func (c *columnSelector) get() *shape.Selector {
	if s := c.selection(); s != nil {
		return s.selector
	}
	return nil
}

func (c *columnSelector) selection() *columnSelection {
	if c == nil {
		return nil
	}
	return c.current.Load()
}

// flattenUnlessColumns runs the --flatten stage only for a document of tabular
// output whose columns no view resolved: the renderer applies the columns of a
// view to the nested document (a flattened document has no nested paths), and
// the table and csv columns of a view are flat already.
func flattenUnlessColumns(flatten output.Stage, cols *columnSelector) output.Stage {
	return output.Map(func(doc jsondoc.JSONDoc) (jsondoc.JSONDoc, error) {
		if cols.get() != nil {
			return doc, nil
		}
		return applyStageToDoc(flatten, doc), nil
	})
}
