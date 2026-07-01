package c8ystream

import (
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/pflag"
	"github.com/tidwall/gjson"
)

// outputTemplateStage builds the --outputTemplate stage. Unlike the SDK's
// context-free template.Jsonnet (which only binds `output` and `index`), this
// routes through mapbuilder.OutputTemplate so the template sees the same
// environment as v1: the `_` helper library, `vars`/`var`, `rand`, `time`,
// `input` (the piped item, as input.value), `flags` (the command's changed
// flags) and `request`/`response` metadata. The per-item input is read from the
// runner, which the producer keeps pointed at the current item.
//
// Returns (nil, nil) when no --outputTemplate is set.
func (r *Runner) outputTemplateStage() (output.Stage, error) {
	tpl := r.Config.GetOutputTemplate()
	if tpl == "" {
		return nil, nil
	}
	ot, err := mapbuilder.GetOutputTemplate(tpl)
	if err != nil {
		return nil, err
	}
	flags := r.commandFlags()
	return output.Map(func(doc jsondoc.JSONDoc) (jsondoc.JSONDoc, error) {
		// input is the piped item that produced this document (nil -> input.value
		// is null, matching v1 for a command run without iteration input).
		var input any
		if len(r.tmplInput) > 0 {
			input = r.tmplInput
		}
		r.tmplMu.Lock()
		reqData, respData := r.tmplRequest, r.tmplResponse
		r.tmplMu.Unlock()
		out, err := ot.Evaluate(doc.Raw(), input, reqData, respData, flags)
		if err != nil {
			return jsondoc.Empty(), err
		}
		return jsondoc.New(unwrapTemplateString(out)), nil
	}), nil
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
