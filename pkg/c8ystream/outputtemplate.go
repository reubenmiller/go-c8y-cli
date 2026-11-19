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

// wrapTemplateJob applies --outputTemplate once per JOB (one input item), not
// once per streamed document — v1 evaluated the template per request against
// the response (a list page's items array), so `devices list --outputTemplate
// input.value` printed one null, not one per device. The job's documents are
// gathered (one doc stays as-is, several become a JSON array, none becomes
// null — v1 ran the template even for empty responses), the template is
// evaluated once with input/request/response bound, and the result is emitted
// as a single document. Errors pass through untouched and suppress the
// evaluation, matching v1's error path.
func (r *Runner) wrapTemplateJob(ot *mapbuilder.OutputTemplate, flags map[string]string, job Job) Job {
	return func(ctx context.Context) output.Seq {
		return func(yield func(jsondoc.JSONDoc, error) bool) {
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
			var body []byte
			switch len(docs) {
			case 0:
				body = []byte("null")
			case 1:
				body = docs[0]
			default:
				body = append(body, '[')
				for i, d := range docs {
					if i > 0 {
						body = append(body, ',')
					}
					body = append(body, d...)
				}
				body = append(body, ']')
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
			out, err := ot.Evaluate(body, input, reqData, respData, flags)
			if err != nil {
				yield(jsondoc.Empty(), err)
				return
			}
			yield(jsondoc.New(unwrapTemplateString(out)), nil)
		}
	}
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
