package c8ystream

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/timestamp"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	ctxhelpers "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/contexthelpers"
)

// Resolver exposes the resolved values of one input item so a command can
// fill the typed arguments of its service call. It does not distinguish query
// parameters from body fields from path parameters — the command assigns each
// value to whatever field of the options struct it belongs to, and go-c8y
// decides how to serialise it.
//
// Resolver is used only while Build runs (on the serial input-draining side);
// the values it returns are captured by the returned Call.
type Resolver struct {
	r       *Runner
	ctx     context.Context
	index   int64
	input   []byte // raw current input item (nil when running without a driver)
	primary string // value resolved by the driver flag (InputFlag), if any
}

// Index is the zero-based position of this item in the input stream.
func (in *Resolver) Index() int64 { return in.index }

// Context returns the run context. Build runs serially, so it is the right
// place to make a resolving API call (e.g. a group name -> id lookup) before
// assembling the typed call.
func (in *Resolver) Context() context.Context { return in.ctx }

// ResolveContext is the run context with dry-run disabled, for reference
// lookups (name -> id) that must hit the real API even under --dry so the
// dry-run report shows the actually-resolved request, matching v1.
func (in *Resolver) ResolveContext() context.Context {
	ctx := apiv2.WithDryRun(in.ctx, false)
	return ctxhelpers.WithMockResponses(ctx, false)
}

// Input returns the raw current input item (the piped line), or nil.
func (in *Resolver) Input() []byte { return in.input }

// String returns a string flag's value, resolving `-`/`-.path` references
// against the current input item so several flags can read one piped object.
// For the driver flag (InputFlag) it returns the value the driver resolved.
func (in *Resolver) String(flag string) string {
	if flag != "" && flag == in.r.primaryName {
		return in.primary
	}
	value, _ := in.r.Cmd.Flags().GetString(flag)
	if path, ok := iterator.InputReference(value); ok {
		return in.ref(path)
	}
	return value
}

// Time resolves a flag holding an absolute or relative timestamp (e.g. "-7d")
// to an ISO-8601 string, or "" when the flag is unset/invalid. Relative-time
// parsing is a CLI concern; go-c8y has no equivalent yet.
func (in *Resolver) Time(flag string) string {
	v := in.String(flag)
	if v == "" {
		return ""
	}
	ts, err := timestamp.TryGetTimestamp(v, false, false)
	if err != nil {
		return ""
	}
	return ts
}

// Bool returns a bool flag's value.
func (in *Resolver) Bool(flag string) bool {
	v, _ := in.r.Cmd.Flags().GetBool(flag)
	return v
}

// StringSlice returns a string-slice flag's value.
func (in *Resolver) StringSlice(flag string) []string {
	v, _ := in.r.Cmd.Flags().GetStringSlice(flag)
	return v
}

// Query evaluates the Cumulocity query expression configured with
// Runner.Query for the current item. The result is decoded once, because the
// flag formatters pre-encode for the v1 raw-query path while the v2 client
// encodes query parameters itself. Empty when no query was configured.
func (in *Resolver) Query() (string, error) {
	if in.r.query == nil {
		return "", nil
	}
	values, _, err := in.r.query.Execute(false)
	if err != nil {
		return "", err
	}
	q, err := url.QueryUnescape(values.Get("q"))
	if err != nil {
		return values.Get("q"), nil
	}
	return q, nil
}

// Body evaluates the body document configured with Runner.Body for the current
// item (templates see the input via input.value). Nil when no body was
// configured.
func (in *Resolver) Body() (json.RawMessage, error) {
	if in.r.body == nil {
		return nil, nil
	}
	var (
		raw []byte
		err error
	)
	if in.input != nil {
		raw, err = in.r.body.MarshalJSONWithInput(in.input)
	} else {
		raw, err = in.r.body.MarshalJSON()
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

// ref extracts a gjson path (empty = whole item) from the current input.
func (in *Resolver) ref(path string) string {
	it := iterator.NewInputRefIterator(in.r.cursor, path, "%s")
	value, _, _ := it.GetNext()
	return string(value)
}

// NameOrID turns a CLI reference into a go-c8y resolver reference: an all-digit
// value is an id (used as-is), anything else is a name to look up ("name:..."),
// matching the smart resolvers in go-c8y. Empty in, empty out.
func NameOrID(value string) string {
	if value == "" {
		return ""
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return "name:" + value
		}
	}
	return value
}
