package c8ystream

import (
	"bytes"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/tidwall/gjson"
)

// bindPipeBodyProperty wires the command's piped value into the request body for
// the generic-pipe driver (Runner.Input).
//
// go-c8y-cli's body machinery reads typed cobra flags, but Runner.Input
// neutralises the legacy per-flag pipe binding so stdin is only read once. That
// leaves the bound flag empty, so the source.id / deviceId / deviceType the
// pipeline maps would never reach the request body and a required-property
// validation would fail (the reason events/alarms/firmware/software create all
// broke on piped input). bindPipeBodyProperty registers a cursor-bound body
// iterator for the mapped property: the body builder evaluates it per item
// (before validation) and only assigns when the property is not already set, so
// an explicit command-line value still wins.
//
// It is a no-op for the per-flag driver (Runner.InputFlag), which already exposes
// the piped value through Resolver.String(primaryName).
func (r *Runner) bindPipeBodyProperty() {
	if r.body == nil || r.primaryName != "" {
		return
	}
	opts, err := flags.GetPipeOptionsFromAnnotation(r.Cmd)
	if err != nil || opts == nil || opts.Property == "" {
		return
	}
	// An explicit command-line value for the bound flag wins over the pipe.
	if opts.Name != "" && r.Cmd.Flags().Changed(opts.Name) {
		return
	}
	r.body.Set(opts.Property, &pipePropertyIterator{cursor: r.cursor, opts: opts})
}

// pipePropertyIterator resolves a command's pipeline property from the shared
// input cursor on each iteration, matching the v1 pipe iterator: from a JSON
// object it plucks the first alias (then the target property) that exists; a
// plain, non-object line is taken whole. It is bound (re-evaluated per item) and
// never ends iteration itself — the input driver owns EOF.
type pipePropertyIterator struct {
	cursor *iterator.InputCursor
	opts   *flags.PipelineOptions
}

func (it *pipePropertyIterator) GetNext() (line []byte, input interface{}, err error) {
	raw := it.cursor.Raw()
	return []byte(pipeBoundValue(raw, it.opts)), raw, nil
}

func (it *pipePropertyIterator) IsBound() bool { return true }

// pipeBoundValue extracts the value a command's pipeline maps to its bound
// property from one input item. A JSON object is plucked by alias (then the
// property); any other line is taken whole. An empty result leaves the body
// property unset.
func pipeBoundValue(input []byte, opts *flags.PipelineOptions) string {
	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return ""
	}
	if jsonUtilities.IsJSONObject(trimmed) {
		props := append(append([]string{}, opts.Aliases...), opts.Property)
		for _, p := range props {
			if p == "" {
				continue
			}
			if v := gjson.GetBytes(trimmed, p); v.Exists() {
				return v.String()
			}
		}
		return ""
	}
	return string(trimmed)
}
