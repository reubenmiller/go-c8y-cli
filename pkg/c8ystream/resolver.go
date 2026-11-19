package c8ystream

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileguard"
	"net/url"
	"strings"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/timestamp"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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
	return apiv2.WithMockResponses(ctx, false)
}

// Input returns the raw current input item (the piped line), or nil.
func (in *Resolver) Input() []byte { return in.input }

// String returns a string flag's value, resolving `-`/`-.path` references
// against the current input item so several flags can read one piped object.
// For the driver flag (InputFlag) it returns the value the driver resolved —
// unless another flag claimed the pipe via `-` (pipeOverride), in which case the
// driver's own flag is left at its (unset) command-line value.
func (in *Resolver) String(flag string) string {
	// The driver flag uses the value the driver resolved, except when another
	// flag claimed the pipe (pipeOverride) AND the driver flag was not given
	// explicitly — then the pipe belongs to the other flag and the driver's own
	// flag stays unset (e.g. `applications list --user -` leaves type empty,
	// while `devices children list --id 12345 --childType -.type` keeps id=12345).
	if flag != "" && flag == in.r.primaryName &&
		(!in.r.pipeOverride || in.r.Cmd.Flags().Changed(flag)) {
		return in.primary
	}
	value, _ := in.r.Cmd.Flags().GetString(flag)
	if path, ok := iterator.InputReference(value); ok {
		return in.refForFlag(flag, path)
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

// TimeValue resolves a flag holding an absolute or relative timestamp to a
// time.Time (zero when unset/invalid). Use it to fill typed time.Time fields of
// an SDK options struct (e.g. events/alarms ListOptions.DateFrom); Time() is
// for string query expressions.
func (in *Resolver) TimeValue(flag string) time.Time {
	s := in.Time(flag)
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.000Z07:00"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// ResolveSourceID resolves a device reference held in the body's "source.id"
// (a name -> id lookup; plain ids pass through) using resolve, returning the
// body with source.id replaced. A body with no non-empty source.id is returned
// unchanged. The CLI owns the name convention (NameOrID); go-c8y performs the
// lookup, and it runs against the real API even under --dry (ResolveContext).
//
// This is the create/update bridge for source-bearing resources (events,
// alarms, measurements): the command builds the full body with the CLI's body
// machinery (--data/--template/typed flags), then this resolves the device
// reference the body carries before the raw create call.
func (in *Resolver) ResolveSourceID(body []byte, resolve func(context.Context, string) (string, error)) (json.RawMessage, error) {
	return in.ResolveBodyRef(body, "source.id", resolve)
}

// ResolveBodyRef resolves a device reference held at the given JSON path in the
// body (name -> id; plain ids pass through) using resolve, returning the body
// with that path replaced. A missing/empty value at path leaves the body
// unchanged. Use it for resources whose device reference is not at source.id
// (e.g. operations: deviceId / agentId). Call once per reference; the result
// chains (each call takes and returns the body). Runs under ResolveContext so it
// hits the real API even under --dry.
func (in *Resolver) ResolveBodyRef(body []byte, path string, resolve func(context.Context, string) (string, error)) (json.RawMessage, error) {
	ref := gjson.GetBytes(body, path).String()
	if ref == "" {
		return json.RawMessage(body), nil
	}
	id, err := resolve(in.ResolveContext(), NameOrID(ref))
	if err != nil {
		return nil, err
	}
	if id == ref {
		return json.RawMessage(body), nil // plain id passed through unchanged
	}
	out, err := sjson.SetBytes(body, path, id)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(out), nil
}

// ApplyOverrideSlice sets path in body to a multi-valued flag's value for the
// current item, cycling through the supplied values one per piped item — the v1
// StringSlice + override-iterator fan-out (e.g. `--device 11111,22222` piped two
// measurements creates one each, source.id 11111 then 22222; a single value
// repeats for every item). A `-`/`-.path` element resolves against the current
// input item. No-op (body returned unchanged) when the flag was not set, so a
// source.id carried by the piped input still wins.
func (in *Resolver) ApplyOverrideSlice(body []byte, flag, path string) (json.RawMessage, error) {
	if !in.r.Cmd.Flags().Changed(flag) {
		return json.RawMessage(body), nil
	}
	values, _ := in.r.Cmd.Flags().GetStringSlice(flag)
	nonEmpty := make([]string, 0, len(values))
	for _, v := range values {
		if p, ok := iterator.InputReference(v); ok {
			v = in.refForFlag(flag, p)
		}
		if v != "" {
			nonEmpty = append(nonEmpty, v)
		}
	}
	if len(nonEmpty) == 0 {
		return json.RawMessage(body), nil
	}
	val := nonEmpty[in.index%int64(len(nonEmpty))]
	out, err := sjson.SetBytes(body, path, val)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(out), nil
}

// ApplyOverrideFirst sets path in body to the flag's first non-empty value
// (FirstString) when the flag was set — the v1 FirstMatch behaviour of a
// secondary reference flag that feeds a body property (e.g. operations create
// --agent → agentId: one agent scopes the whole batch, extra values are
// ignored). No-op (body returned unchanged) when the flag is unset or empty,
// so a value carried by --data/--template or the piped input still wins.
func (in *Resolver) ApplyOverrideFirst(body []byte, flag, path string) (json.RawMessage, error) {
	if !in.r.Cmd.Flags().Changed(flag) {
		return json.RawMessage(body), nil
	}
	v := in.FirstString(flag)
	if v == "" {
		return json.RawMessage(body), nil
	}
	out, err := sjson.SetBytes(body, path, v)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(out), nil
}

// Bool returns a bool flag's value.
func (in *Resolver) Bool(flag string) bool {
	v, _ := in.r.Cmd.Flags().GetBool(flag)
	return v
}

// BoolPtr returns a pointer to a bool flag's value. It is never nil: the flag
// has a registered default, so the value is always meaningful and is always
// sent. Use it for query parameters that must carry an explicit false (e.g.
// forceCascade with a default of true) rather than being dropped —
// go-querystring's omitempty omits a nil pointer but keeps a pointer to false,
// which a plain bool cannot express.
func (in *Resolver) BoolPtr(flag string) *bool {
	v, _ := in.r.Cmd.Flags().GetBool(flag)
	return &v
}

// BoolPtrIfChanged returns a pointer to a bool flag's value only when the user
// set the flag, and nil otherwise. Use it for query parameters whose default is
// "off" so an unset flag stays off the request entirely, while an explicit
// value (including false) is sent — matching the v1 behaviour of only emitting
// the parameter when the user supplied it.
func (in *Resolver) BoolPtrIfChanged(flag string) *bool {
	if !in.r.Cmd.Flags().Changed(flag) {
		return nil
	}
	v, _ := in.r.Cmd.Flags().GetBool(flag)
	return &v
}

// FirstString returns the first non-empty value of a flag, resolving
// `-`/`-.path` references against the current input item. For a slice flag
// this is the v1 "FirstMatch" semantics of secondary reference flags: a
// scoping lookup (e.g. `firmware versions get --firmware pkg`) takes exactly
// one reference, and extra values are ignored — matching v1's
// WithReferenceByNameFirstMatch, which resolved values[0]. For a plain string
// flag it behaves like String.
func (in *Resolver) FirstString(flag string) string {
	f := in.r.Cmd.Flags().Lookup(flag)
	if f == nil {
		return ""
	}
	if !strings.Contains(f.Value.Type(), "Slice") {
		return in.String(flag)
	}
	for _, v := range in.StringSlice(flag) {
		if v != "" {
			return v
		}
	}
	return ""
}

// StringSlice returns a string-slice flag's value, resolving any `-`/`-.path`
// element against the current input item (e.g. `--user -`).
func (in *Resolver) StringSlice(flag string) []string {
	v, _ := in.r.Cmd.Flags().GetStringSlice(flag)
	for i, item := range v {
		if path, ok := iterator.InputReference(item); ok {
			v[i] = in.refForFlag(flag, path)
		}
	}
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

// refForFlag resolves a `-`/`-.path` flag reference against the current input
// item, matching the v1 pipe-binding extraction:
//   - bare `-`: a JSON object yields the flag's own property; a plain line is
//     taken whole.
//   - `-.a,.b`: the first existing of the comma-separated gjson paths (leading
//     dots optional), so a fallback chain like `-.id,.one` resolves to whichever
//     property the piped object carries.
func (in *Resolver) refForFlag(flag, path string) string {
	raw := bytes.TrimSpace(in.r.cursor.Raw())
	if path == "" {
		if jsonUtilities.IsJSONObject(raw) {
			return gjsonFirst(raw, []string{flag})
		}
		return string(raw)
	}
	return gjsonFirst(raw, splitRefPaths(path))
}

// splitRefPaths turns a `-.a,.b` reference tail into the gjson paths ["a","b"],
// trimming the optional leading dot of each fallback segment.
func splitRefPaths(path string) []string {
	parts := strings.Split(path, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.TrimPrefix(p, ".")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// gjsonFirst returns the first of paths that exists in raw, or "".
func gjsonFirst(raw []byte, paths []string) string {
	for _, p := range paths {
		if v := gjson.GetBytes(raw, p); v.Exists() {
			return v.String()
		}
	}
	return ""
}

// NameOrID turns a CLI reference into a go-c8y resolver reference: an all-digit
// value is an id (used as-is), anything else is a name to look up ("name:..."),
// matching the smart resolvers in go-c8y. Empty in, empty out.
//
// A comma-joined value is split and the first id-looking part wins, falling
// back to the first non-empty part as a name — the v1 fetcher semantics
// (parseAndSanitizeIDs + FirstMatch), which let a piped csv line like
// "83001,1.0.0" (id,version) resolve to the id.
func NameOrID(value string) string {
	if value == "" {
		return ""
	}
	if strings.Contains(value, ",") {
		parts := strings.Split(strings.ReplaceAll(value, ", ", ","), ",")
		name := ""
		for _, p := range parts {
			if p == "" {
				continue
			}
			if isDigits(p) {
				return p
			}
			if name == "" {
				name = p
			}
		}
		if name == "" {
			return ""
		}
		return "name:" + name
	}
	if !isDigits(value) {
		return "name:" + value
	}
	return value
}

// FirstID returns the first comma-separated all-digit part of a reference
// value — the v1 fetcher semantics that let a piped csv line like
// "83001,1.0.0" (id,version) resolve to the id — or the value unchanged when
// no part looks like an id. Use it for reference builders that bypass
// NameOrID (e.g. version/patch refs that take the raw value).
func FirstID(value string) string {
	if !strings.Contains(value, ",") {
		return value
	}
	for _, p := range strings.Split(strings.ReplaceAll(value, ", ", ","), ",") {
		if isDigits(p) {
			return p
		}
	}
	return value
}

// isDigits reports whether s is non-empty and all ASCII digits (an id).
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// File returns a file-path flag's value after checking it against the denied
// paths (C8Y_DENIED_PATHS / files.deniedPaths). Use it for every flag whose
// value is handed to the SDK as a path to read or upload, so a command cannot
// bypass the file guard by letting go-c8y open the file itself.
func (in *Resolver) File(flag string) (string, error) {
	path := in.String(flag)
	if err := fileguard.Check(path); err != nil {
		return "", err
	}
	return path, nil
}
