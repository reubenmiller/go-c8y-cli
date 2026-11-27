package c8ystream

import (
	"context"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/pflag"
	"github.com/tidwall/gjson"
)

// prepareOutputTemplate parses --outputTemplate once for the run. Unlike the
// SDK's context-free template.Jsonnet (which only binds `output` and `index`),
// evaluation routes through mapbuilder.OutputTemplate so the template sees the
// same environment as v1: the `_` helper library, `vars`/`var`, `rand`,
// `time`, `input` (the piped item, as input.value), `flags` (the command's
// changed flags) and `request`/`response` metadata.
//
// Returns (nil, nil) when no --outputTemplate is set.
func (r *Runner) prepareOutputTemplate() (*mapbuilder.OutputTemplate, error) {
	tpl := r.Config.GetOutputTemplate()
	if tpl == "" {
		return nil, nil
	}
	return mapbuilder.GetOutputTemplate(tpl)
}

// wrapTemplateJob applies --outputTemplate once per JOB (one input item) and
// per response, as v1 and v2.59.0 did: the template is evaluated against the
// response, not against each streamed document, so `devices list
// --outputTemplate input.value` prints one null, not one per device.
//
// What `output` is, as in v2.59.0:
//   - a list (the response has a collection property or is an array): the
//     array of the items of the page, also when it holds one item or none
//     ("[]");
//   - --raw, --withTotalPages or --withTotalElements: the whole response, with
//     its envelope and statistics;
//   - anything else: the document (null when there is none).
//
// A template which returns an array prints one document per element (as
// v2.59.0), except under --raw, --withTotalPages and --withTotalElements,
// where the array is printed as a whole. Errors pass through untouched and
// suppress the evaluation, matching v1's error path.
func (r *Runner) wrapTemplateJob(ot *mapbuilder.OutputTemplate, flags map[string]string, job Job) Job {
	whole := r.Config.RawOutput()
	if common, err := r.Config.GetOutputCommonOptions(r.Cmd); err == nil {
		whole = whole || common.WithTotalPages || common.WithTotalElements
	}
	// the elements of an array result are printed one by one, except when the
	// whole response is printed (a top-level array is still shaped item by item
	// with --raw --select, as in v2.59.0)
	split := !whole || (r.Config.RawOutput() && r.shapesItems())

	return func(ctx context.Context) output.Seq {
		return func(yield func(jsondoc.JSONDoc, error) bool) {
			r.tmplMu.Lock()
			r.tmplPages = nil
			r.tmplMu.Unlock()

			var docs [][]byte
			failed := false
			job(ctx)(func(doc jsondoc.JSONDoc, err error) bool {
				if err != nil {
					failed = true
					return yield(doc, err)
				}
				docs = append(docs, doc.Raw())
				return true
			})
			if failed {
				return
			}
			// input is the piped item that produced this job (nil -> input.value
			// is null, matching v1 for a command run without iteration input).
			var input any
			if len(r.tmplInput) > 0 {
				input = r.tmplInput
			}
			r.tmplMu.Lock()
			reqData, respData := r.tmplRequest, r.tmplResponse
			r.tmplMu.Unlock()

			for _, body := range r.templateBodies(docs, whole) {
				out, err := ot.Evaluate(body, input, reqData, respData, flags)
				if err != nil {
					yield(jsondoc.Empty(), err)
					return
				}
				out = unwrapTemplateString(out)
				if res := gjson.ParseBytes(out); split && res.IsArray() {
					for _, item := range res.Array() {
						if !yield(jsondoc.New([]byte(item.Raw)), nil) {
							return
						}
					}
					continue
				}
				if !yield(jsondoc.New(out), nil) {
					return
				}
			}
		}
	}
}

// templatePage is a response of the current job, which the --outputTemplate
// evaluation reads to tell what `output` is (see templateBodies).
type templatePage struct {
	method, path string
	body         []byte
}

// jobPages returns the JSON responses of the pages of the current job: the
// ones with the method and path of the last response (the requests which
// resolved names to ids are left out).
func (r *Runner) jobPages() []templatePage {
	r.tmplMu.Lock()
	defer r.tmplMu.Unlock()
	if len(r.tmplPages) == 0 {
		return nil
	}
	last := r.tmplPages[len(r.tmplPages)-1]
	var pages []templatePage
	for _, p := range r.tmplPages {
		if p.method == last.method && p.path == last.path && gjson.ValidBytes(p.body) {
			pages = append(pages, p)
		}
	}
	return pages
}

// templateBodies returns the inputs (`output`) which the template is
// evaluated with, one evaluation each: see wrapTemplateJob. docs are the
// documents the job produced.
func (r *Runner) templateBodies(docs [][]byte, whole bool) [][]byte {
	pages := r.jobPages()

	if whole {
		// the whole response of each page, with the envelope and statistics
		if len(pages) > 0 {
			bodies := make([][]byte, 0, len(pages))
			for _, p := range pages {
				bodies = append(bodies, p.body)
			}
			return bodies
		}
		return [][]byte{joinDocs(docs)}
	}

	// A list: the items of each page as an array, which is also an empty array
	// for an empty page. Only when the items add up to the documents of the
	// job (the command did not drop or add any).
	if len(pages) > 0 {
		var bodies [][]byte
		total := 0
		isList := true
		for _, p := range pages {
			items, ok := r.pageItems(p.body)
			if !ok {
				isList = false
				break
			}
			total += len(gjson.ParseBytes(items).Array())
			bodies = append(bodies, items)
		}
		if isList && total == len(docs) {
			return bodies
		}
	}
	return [][]byte{joinDocs(docs)}
}

// pageItems returns the array of the items of a response: the response itself
// when it is an array, or else its collection property.
func (r *Runner) pageItems(body []byte) ([]byte, bool) {
	if res := gjson.ParseBytes(body); res.IsArray() {
		return []byte(res.Raw), true
	}
	property := r.resultProperty(body)
	if property == "" {
		return nil, false
	}
	items := gjson.GetBytes(body, property)
	if !items.IsArray() {
		return nil, false
	}
	return []byte(items.Raw), true
}

// joinDocs is the template input of the documents of a job without a list
// response: one document as is, several as a JSON array, none as null.
func joinDocs(docs [][]byte) []byte {
	switch len(docs) {
	case 0:
		return []byte("null")
	case 1:
		return docs[0]
	}
	body := []byte{'['}
	for i, d := range docs {
		if i > 0 {
			body = append(body, ',')
		}
		body = append(body, d...)
	}
	return append(body, ']')
}

// commandFlags captures the command's explicitly-set flags as the `flags`
// template variable, matching v1's CommonCommandOptions.CommandFlags (the
// "[]" wrapping a slice prints is trimmed).
func (r *Runner) commandFlags() map[string]string {
	out := make(map[string]string)
	r.Cmd.Flags().Visit(func(f *pflag.Flag) {
		out[f.Name] = strings.Trim(f.Value.String(), "[]")
	})
	return out
}

// unwrapTemplateString reproduces v1's output handling for a template whose
// result is a top-level JSON string (e.g. std.join(',', …) or
// std.manifestTomlEx(…)): the decoded string is emitted verbatim rather than as
// a quoted/escaped JSON string. Numbers, bools, null, objects and arrays pass
// through unchanged.
func unwrapTemplateString(out []byte) []byte {
	res := gjson.ParseBytes(out)
	if res.Type == gjson.String {
		return []byte(res.Str)
	}
	return out
}
