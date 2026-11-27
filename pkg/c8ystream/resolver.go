package c8ystream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileguard"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/source"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/timestamp"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
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

	// lastFlag is the flag which was read last. A lookup which fails while
	// Build runs is the lookup of the value of this flag (a command reads the
	// flag, then resolves it), see lookupFlagIsSecondary.
	lastFlag string
}

// lookupFlagIsSecondary reports whether the lookup which failed with err was
// the lookup of a flag which does not drive the command (InputFlag), e.g.
// --device of "remoteaccess configurations get", which is driven by --id: the
// last flag read is another flag, and the name which could not be looked up
// is not the value of the flag which drives the command (a command may read
// that value, then other flags, then look it up).
func (in *Resolver) lookupFlagIsSecondary(err error) bool {
	if in.lastFlag == "" || in.lastFlag == in.r.primaryName {
		return false
	}
	if name, ok := lookupName(err); ok && in.primary != "" && name == in.primary {
		return false
	}
	return true
}

// lookupName returns the name of a failed lookup: the name which matched
// nothing, or the reference which could not be resolved (without "name:").
func lookupName(err error) (string, bool) {
	if _, name, ok := core.NameNotFound(err); ok {
		return name, true
	}
	var lookupErr *core.LookupError
	if errors.As(err, &lookupErr) {
		return strings.TrimPrefix(lookupErr.Reference, "name:"), true
	}
	return "", false
}

// Index is the zero-based position of this item in the input stream.
func (in *Resolver) Index() int64 { return in.index }

// Context returns the run context. Build runs serially, so it is the right
// place to make a resolving API call (e.g. a group name -> id lookup) before
// assembling the typed call.
func (in *Resolver) Context() context.Context { return in.ctx }

// ResolveContext is the run context prepared for a reference lookup that a
// command makes itself (name -> id): the SDK's own resolution context, so the
// lookup hits the real API even under --dry and keeps the session's request
// options. Prefer passing the reference to the SDK method, which resolves it.
func (in *Resolver) ResolveContext() context.Context {
	return apiv2.ResolutionContext(in.ctx)
}

// ResolveContextFor is ResolveContext for the lookup of the value of a flag.
// Name the flag so a lookup which fails is reported the way v2.59.0 reported
// it: a flag which does not drive the command (not the InputFlag) has its own
// text and exit code (see LookupTransportError and NoMatchesFlagError).
func (in *Resolver) ResolveContextFor(flag string) context.Context {
	in.lastFlag = flag
	return in.ResolveContext()
}

// Input returns the raw current input item (the piped line), or nil.
func (in *Resolver) Input() []byte { return in.input }

// String returns a string flag's value, resolving `-`/`-.path` references
// against the current input item so several flags can read one piped object.
// For the driver flag (InputFlag) it returns the value the driver resolved —
// unless another flag claimed the pipe via `-` (pipeOverride), in which case the
// driver's own flag is left at its (unset) command-line value.
func (in *Resolver) String(flag string) string {
	in.lastFlag = flag
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
	in.lastFlag = flag
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
	in.lastFlag = flag
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

// RawBody returns the body configured with Runner.Body when it is raw text
// (e.g. --data with a non-JSON content type) rather than a JSON document.
func (in *Resolver) RawBody() (string, bool) {
	if in.r.body == nil || !in.r.body.HasRaw() {
		return "", false
	}
	return in.r.body.GetRaw(), true
}

// BodyFile returns the file which is sent as the request body (e.g. a binary upload), or nil
// when the body is not set from a file.
func (in *Resolver) BodyFile() (*os.File, error) {
	if in.r.body == nil || !in.r.body.HasFile() {
		return nil, nil
	}
	return in.r.body.GetFile()
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

// NameOrID turns a CLI reference into a go-c8y resolver reference using the
// SDK convention (source.ParseRef): all digits is an id, a scheme prefix
// (name:, ext:, query:, id:) is kept, anything else is a name. A comma list
// picks its first id, else its first part as a name. Empty in, empty out.
func NameOrID(value string) string {
	return source.ParseRef(value)
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
