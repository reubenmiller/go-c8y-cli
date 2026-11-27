package c8ystream

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newErrorRunner(withError bool) *Runner {
	v := viper.New()
	v.Set("settings.defaults.withError", withError)
	return &Runner{Config: config.NewConfig(v), Factory: &cmdutil.Factory{}}
}

func serverErr(method, rawURL string, status int, body string) *core.Error {
	u, _ := url.Parse(rawURL)
	return &core.Error{
		Code:       status,
		Type:       "inventory/Not Found",
		Message:    "No managedObject for id '1'! https://cumulocity.com/api/core/",
		MessageRaw: body,
		Response:   &http.Response{Request: &http.Request{Method: method, URL: u}},
	}
}

const notFoundBody = `{"error":"inventory/Not Found","message":"No managedObject for id '1'!","info":"https://cumulocity.com/api/core/"}`

func TestToServerError_v1Format(t *testing.T) {
	r := newErrorRunner(false)
	got := r.toServerError(serverErr("GET", "http://host/inventory/managedObjects/1?a=b", 404, notFoundBody))
	ce, ok := got.(cmderrors.CommandError)
	if !ok {
		t.Fatalf("want a CommandError, got %T", got)
	}
	if ce.ErrorType != cmderrors.ErrTypeServer || ce.ExitCode != cmderrors.ExitNotFound404 || ce.StatusCode != 404 {
		t.Errorf("unexpected type/exit/status: %+v", ce)
	}
	// the message is the server message without the info url
	if ce.Message != "No managedObject for id '1'!" {
		t.Errorf("message = %q", ce.Message)
	}
	want := "serverError: No managedObject for id '1'! GET http://host/inventory/managedObjects/1?a=b: 404 inventory/Not Found No managedObject for id '1'!"
	if ce.Error() != want {
		t.Errorf("Error() = %q, want %q", ce.Error(), want)
	}
	// by default the raw response is printed below the message
	if !ce.WithRawMessage {
		t.Error("the response block must be printed without --withError")
	}
	// the error is still the SDK error for status checks
	if !core.IsNotFound(ce) {
		t.Error("the SDK error must be reachable through Unwrap")
	}
}

func TestToServerError_withErrorJSON(t *testing.T) {
	r := newErrorRunner(true)
	ce := r.toServerError(serverErr("GET", "http://host/inventory/managedObjects/1", 404, notFoundBody)).(cmderrors.CommandError)
	if ce.WithRawMessage {
		t.Error("--withError prints only the JSON, not the response block")
	}
	var doc struct {
		Error map[string]any `json:"error"`
	}
	if err := json.Unmarshal([]byte(ce.JSONString()), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Error) != 3 || doc.Error["error"] != "inventory/Not Found" ||
		doc.Error["message"] != "No managedObject for id '1'!" || doc.Error["info"] != "https://cumulocity.com/api/core/" {
		t.Errorf("error = %v, want {error, message, info}", doc.Error)
	}
}

func TestToServerError_unresolvedName(t *testing.T) {
	r := newErrorRunner(false)
	tests := map[string]error{
		"nosuch":   core.NewNameNotFoundError("managed object", "nosuch", "managed object not found with name: %s", "nosuch"),
		"my app":   core.NewNameNotFoundError("application", "my app", "application not found with name: %s, type: %s", "my app", "MICROSERVICE"),
		"nogroup":  serverErr("GET", "http://host/user/t1/groupByName/nogroup", 404, `{"error":"user/group/notFound","message":"Not found"}`),
		"someName": core.WrapLookup("microservice", "name:someName", core.NewNameNotFoundError("microservice", "someName", "microservice not found with name: %s", "someName")),
		"fw":       core.WrapLookup("firmware", "name:fw", core.NewNameNotFoundError("firmware", "fw", "firmware not found: name=%s", "fw")),
		"cfg":      core.NewNameNotFoundError("configuration", "cfg", "configuration not found: %s", "cfg"),
	}
	for name, in := range tests {
		ce, ok := r.toServerError(in).(cmderrors.CommandError)
		if !ok {
			t.Fatalf("%s: want a CommandError", name)
		}
		want := "commandError: referenceByName: no matching items found. name=" + name + "  ::StatusCode=404"
		if ce.Error() != want || ce.ExitCode != cmderrors.ExitNotFound404 || ce.StatusCode != 404 {
			t.Errorf("%s: got %q exit=%d status=%d, want %q exit=4", name, ce.Error(), ce.ExitCode, ce.StatusCode, want)
		}
	}
	// a 404 of an id is not a name lookup
	if ce := r.toServerError(serverErr("GET", "http://host/inventory/managedObjects/1", 404, notFoundBody)).(cmderrors.CommandError); ce.ErrorType != cmderrors.ErrTypeServer {
		t.Errorf("a 404 of an id must stay a server error, got %s", ce.ErrorType)
	}
	// the 404 of usergroups getByName (the command's own request) is a server error
	r.DirectNameLookup()
	if ce := r.toServerError(serverErr("GET", "http://host/user/t1/groupByName/nogroup", 404, `{"error":"user/group/notFound","message":"Not found"}`)).(cmderrors.CommandError); ce.ErrorType != cmderrors.ErrTypeServer || ce.ExitCode != cmderrors.ExitNotFound404 {
		t.Errorf("a direct group name lookup must stay a server error, got %s exit=%d", ce.ErrorType, ce.ExitCode)
	}
}

func TestToServerError_bootstrapUserIsUserError(t *testing.T) {
	r := newErrorRunner(false)
	in := serverErr("GET", "http://host/application/applications/9/bootstrapUser", 404, `{"error":"application/notFound","message":"Not found"}`)
	// without UserErrorRequests (microservices getBootstrapUser) it is a server error
	if ce := r.toServerError(in).(cmderrors.CommandError); ce.ErrorType != cmderrors.ErrTypeServer || ce.ExitCode != cmderrors.ExitNotFound404 {
		t.Errorf("got exit=%d type=%s, want 4 serverError", ce.ExitCode, ce.ErrorType)
	}
	r.UserErrorRequests("/bootstrapUser")
	got := r.toServerError(in)
	ce := got.(cmderrors.CommandError)
	if ce.ExitCode != cmderrors.ExitUserError || ce.ErrorType != cmderrors.ErrTypeCommand {
		t.Errorf("got exit=%d type=%s, want 101 commandError", ce.ExitCode, ce.ErrorType)
	}
	if want := "commandError: GET http://host/application/applications/9/bootstrapUser: 404 application/notFound Not found"; ce.Error() != want {
		t.Errorf("Error() = %q, want %q", ce.Error(), want)
	}
}

func TestWaitRequestError(t *testing.T) {
	got := WaitRequestError(serverErr("GET", "http://host/devicecontrol/operations/x", 404, `{"error":"devicecontrol/Not Found","message":"No operation for id 'x'!"}`))
	ce, ok := got.(cmderrors.CommandError)
	if !ok || ce.ExitCode != cmderrors.ExitError || ce.ErrorType != cmderrors.ErrTypeCommand {
		t.Fatalf("got %#v, want a commandError with exit code 100", got)
	}
	if want := "GET http://host/devicecontrol/operations/x: 404 devicecontrol/Not Found No operation for id 'x'!"; ce.Message != want {
		t.Errorf("message = %q, want %q", ce.Message, want)
	}
	plain := errors.New("boom")
	if WaitRequestError(plain) != plain || WaitRequestError(nil) != nil {
		t.Error("other errors pass through")
	}
}

func TestCollapseErrors_assertAggregate(t *testing.T) {
	r := newErrorRunner(false)
	errs := []error{errors.New("a"), errors.New("b")}

	plain := r.collapseErrors(errs).(cmderrors.CommandError)
	if plain.ExitCode != cmderrors.ExitCompletedWithErrors || plain.Processed {
		t.Errorf("a command reports the summary line with exit 104, got %+v", plain)
	}

	r.Assert(true)
	agg := r.collapseErrors(errs).(cmderrors.CommandError)
	if agg.ExitCode != cmderrors.ExitCompletedWithErrors || !agg.Processed {
		t.Errorf("an assert command has no summary line, got %+v", agg)
	}
	// the batch error keeps exit 104 and is not turned into an assertion error
	out := r.assert.exitCode(agg).(cmderrors.CommandError)
	if out.ExitCode != cmderrors.ExitCompletedWithErrors {
		t.Errorf("exit = %d, want 104", out.ExitCode)
	}
	single := r.assert.exitCode(errors.New("x")).(cmderrors.CommandError)
	if single.ExitCode != cmderrors.ExitAssertionError {
		t.Errorf("exit = %d, want 112", single.ExitCode)
	}
}

func TestToServerError_assertRequestIsCommandError(t *testing.T) {
	r := newErrorRunner(false)
	r.Assert(true)
	ce := r.toServerError(serverErr("GET", "http://host/inventory/managedObjects/1", 404, notFoundBody)).(cmderrors.CommandError)
	if ce.ErrorType != cmderrors.ErrTypeCommand || ce.CumulocityError != nil || ce.WithRawMessage {
		t.Errorf("an assert command reports a plain command error, got %+v", ce)
	}
	want := "commandError: GET http://host/inventory/managedObjects/1: 404 inventory/Not Found No managedObject for id '1'!"
	if ce.Error() != want {
		t.Errorf("Error() = %q, want %q", ce.Error(), want)
	}
}

// transportError is the error of a request which got no response, as the SDK
// returns it (the Go error of the http client wrapped in a core.Error)
func transportError(t *testing.T) error {
	t.Helper()
	urlErr := &url.Error{Op: "Get", URL: "http://127.0.0.1:1/inventory/managedObjects/1", Err: errors.New("dial tcp 127.0.0.1:1: connect: connection refused")}
	return &core.Error{Code: core.ErrorFromError, Message: urlErr.Error(), Err: urlErr}
}

func TestToServerError_transportError(t *testing.T) {
	r := newErrorRunner(false)
	got := r.toServerError(transportError(t))
	ce, ok := got.(cmderrors.CommandError)
	if !ok {
		t.Fatalf("want a CommandError, got %T", got)
	}
	// v2.59.0: serverError, exit code 99, the message of the http client twice
	if ce.ErrorType != cmderrors.ErrTypeServer || ce.ExitCode != cmderrors.ExitUnknownError {
		t.Errorf("unexpected type/exit: %+v", ce)
	}
	msg := `Get "http://127.0.0.1:1/inventory/managedObjects/1": dial tcp 127.0.0.1:1: connect: connection refused`
	if want := "serverError: " + msg + " " + msg; ce.Error() != want {
		t.Errorf("Error() = %q, want %q", ce.Error(), want)
	}

	// the --withError document has the error of the http client under "error"
	raw, err := json.Marshal(ce)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		ErrorType string `json:"errorType"`
		Message   string `json:"message"`
		ExitCode  int    `json:"exitCode"`
		Error     struct {
			Op  string
			URL string
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ErrorType != "serverError" || doc.Message != msg || doc.ExitCode != 99 || doc.Error.Op != "Get" || doc.Error.URL == "" {
		t.Errorf("unexpected document: %s", raw)
	}
}

func TestToServerError_transportErrorOfAssert(t *testing.T) {
	r := newErrorRunner(false)
	r.assert = &assertion{}
	ce := r.toServerError(transportError(t)).(cmderrors.CommandError)
	if ce.ErrorType != cmderrors.ErrTypeCommand || ce.ExitCode != cmderrors.ExitError {
		t.Errorf("an assert command reports a plain command error: %+v", ce)
	}
}

func TestWaitRequestError_transportError(t *testing.T) {
	ce, ok := WaitRequestError(transportError(t)).(cmderrors.CommandError)
	if !ok || ce.ErrorType != cmderrors.ErrTypeCommand || ce.ExitCode != cmderrors.ExitError {
		t.Errorf("a wait command reports a command error with exit code 100: %+v", ce)
	}
}

func TestTransportError_notATransportError(t *testing.T) {
	for name, err := range map[string]error{
		"plain error":      errors.New("boom"),
		"server error":     serverErr("GET", "http://host/x", 500, "{}"),
		"cancelled":        &url.Error{Op: "Get", URL: "http://host/x", Err: context.Canceled},
		"cancelled in sdk": &core.Error{Code: core.ErrorFromError, Err: &url.Error{Op: "Get", URL: "http://host/x", Err: context.Canceled}},
	} {
		if _, ok := TransportError(err); ok {
			t.Errorf("%s must not be a transport error", name)
		}
	}
}

// TestToServerError_unresolvedNameIsNotMatchedByText checks that the CLI relies
// on the typed error of the SDK: a 404 which only says so in its text is a
// normal server error.
func TestToServerError_unresolvedNameIsNotMatchedByText(t *testing.T) {
	r := newErrorRunner(false)
	for _, in := range []error{
		core.ErrNotFound("managed object not found with name: %s", "nosuch"),
		core.ErrNotFound("remote access configuration not found: name=%s", "nosuch"),
	} {
		ce, ok := r.toServerError(in).(cmderrors.CommandError)
		if !ok || ce.ErrorType != cmderrors.ErrTypeServer {
			t.Errorf("%v: want a serverError, got %#v", in, r.toServerError(in))
		}
	}
	if IsUnresolvedName(core.ErrNotFound("managed object not found with name: x")) {
		t.Error("a text must not make an unresolved name")
	}
	if !IsUnresolvedName(core.WrapLookup("managed object", "name:x", core.NewNameNotFoundError("managed object", "x", "x"))) {
		t.Error("a name not found error is an unresolved name")
	}
}

func refused() error {
	return &url.Error{Op: "Get", URL: "http://127.0.0.1:1/inventory/managedObjects?pageSize=1", Err: errors.New("dial tcp 127.0.0.1:1: connect: connection refused")}
}

// TestToServerError_transportErrorOfLookup checks the error of a name lookup
// which got no response, per kind of resource (the texts of v2.59.0).
func TestToServerError_transportErrorOfLookup(t *testing.T) {
	r := newErrorRunner(false)
	tests := []struct {
		resource string
		want     string
	}{
		{"managed object", "commandError: Could not fetch by id: "},
		{"firmware", "commandError: Could not fetch by id: "},
		{"user group", "commandError: Could not fetch by id: "},
		{"application", "commandError: Get "},
		{"microservice", "commandError: could not fetch microservices: "},
		{"ui plugin", "commandError: could not fetch ui plugin: "},
		{"hosted application", "commandError: could not fetch applications: "},
		{"remote access configuration", "commandError: could not fetch microservices: "},
	}
	for _, tc := range tests {
		got := r.toServerError(core.WrapLookup(tc.resource, "name:x", refused()))
		ce, ok := got.(cmderrors.CommandError)
		if !ok || ce.ErrorType != cmderrors.ErrTypeCommand || ce.ExitCode != cmderrors.ExitUserError {
			t.Errorf("%s: want a commandError with exit code 101, got %#v", tc.resource, got)
			continue
		}
		want := tc.want + refused().Error()
		if tc.resource == "application" {
			want = "commandError: " + refused().Error()
		}
		if ce.Error() != want {
			t.Errorf("%s: Error() = %q, want %q", tc.resource, ce.Error(), want)
		}
	}

	// the transport error of the request of the command is a server error
	ce := r.toServerError(refused()).(cmderrors.CommandError)
	if ce.ErrorType != cmderrors.ErrTypeServer || ce.ExitCode != cmderrors.ExitUnknownError {
		t.Errorf("own request: got %s exit=%d, want serverError 99", ce.ErrorType, ce.ExitCode)
	}

	// an HTTP error of the lookup request is not a transport error
	ce = r.toServerError(core.WrapLookup("managed object", "name:x", serverErr("GET", "http://host/inventory/managedObjects", 500, `{"error":"x","message":"boom"}`))).(cmderrors.CommandError)
	if ce.ErrorType != cmderrors.ErrTypeServer || ce.ExitCode != cmderrors.ExitInternalServerError500 {
		t.Errorf("http error of a lookup: got %s exit=%d, want a serverError", ce.ErrorType, ce.ExitCode)
	}
}

func TestLookupTransportError_secondaryFlag(t *testing.T) {
	ce, ok := LookupTransportError(core.WrapLookup("managed object", "name:x", refused()), true)
	if !ok || ce.ExitCode != cmderrors.ExitError {
		t.Fatalf("got %#v ok=%v, want exit code 100", ce, ok)
	}
	if want := "commandError: failed to lookup by name. Could not fetch by id: " + refused().Error(); ce.Error() != want {
		t.Errorf("Error() = %q, want %q", ce.Error(), want)
	}
	if _, ok := LookupTransportError(refused(), true); ok {
		t.Error("a transport error of the request itself is no lookup error")
	}
	if _, ok := LookupTransportError(core.WrapLookup("managed object", "x", errors.New("boom")), false); ok {
		t.Error("only a transport error is turned into the lookup error")
	}
}

func TestRunner_lookupFailureStyle(t *testing.T) {
	r := newErrorRunner(false)
	r.LookupFailureStyle("failed to find hosted application. ", cmderrors.ExitError)
	ce := r.toServerError(core.WrapLookup("hosted application", "x", refused())).(cmderrors.CommandError)
	if ce.ExitCode != cmderrors.ExitError {
		t.Errorf("exit = %d, want 100", ce.ExitCode)
	}
	if want := "commandError: failed to find hosted application. could not fetch applications: " + refused().Error(); ce.Error() != want {
		t.Errorf("Error() = %q, want %q", ce.Error(), want)
	}
}

func TestNoMatchesFlagError(t *testing.T) {
	ce := NoMatchesFlagError().(cmderrors.CommandError)
	if ce.Error() != "commandError: reference by name: no matches found" || ce.ExitCode != cmderrors.ExitError {
		t.Errorf("got %q exit=%d", ce.Error(), ce.ExitCode)
	}
}

// TestRunner_buildErrorOfLookup checks the error of a name which is looked up
// while the call is built: it depends on the flag which holds the name.
func TestRunner_buildErrorOfLookup(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringSlice("id", nil, "")
	cmd.Flags().StringSlice("device", nil, "")
	r := newErrorRunner(false)
	r.Cmd = cmd
	r.primaryName = "id"
	in := &Resolver{r: r, ctx: context.Background()}

	notFound := core.WrapLookup("managed object", "name:x", core.NewNameNotFoundError("managed object", "x", "managed object not found with name: x"))
	lookup := core.WrapLookup("managed object", "name:x", refused())

	// the flag which drives the command
	in.ResolveContextFor("id")
	if ce := r.buildError(in, lookup).(cmderrors.CommandError); ce.ExitCode != cmderrors.ExitUserError || ce.Message != "Could not fetch by id: "+refused().Error() {
		t.Errorf("driver flag, transport: got %q exit=%d", ce.Message, ce.ExitCode)
	}
	if got := r.buildError(in, notFound); got != notFound {
		t.Errorf("driver flag, not found: got %v, want the error unchanged (reported as an unresolved name later)", got)
	}

	// another flag
	_ = in.FirstString("device")
	if ce := r.buildError(in, lookup).(cmderrors.CommandError); ce.ExitCode != cmderrors.ExitError || ce.Message != "failed to lookup by name. Could not fetch by id: "+refused().Error() {
		t.Errorf("secondary flag, transport: got %q exit=%d", ce.Message, ce.ExitCode)
	}
	if ce := r.buildError(in, notFound).(cmderrors.CommandError); ce.Error() != "commandError: reference by name: no matches found" || ce.ExitCode != cmderrors.ExitError {
		t.Errorf("secondary flag, not found: got %q exit=%d", ce.Error(), ce.ExitCode)
	}

	// the value of the flag which drives the command, looked up after another
	// flag was read
	in.primary = "x"
	_ = in.FirstString("device")
	if ce := r.buildError(in, lookup).(cmderrors.CommandError); ce.ExitCode != cmderrors.ExitUserError || ce.Message != "Could not fetch by id: "+refused().Error() {
		t.Errorf("driver flag read before another flag, transport: got %q exit=%d", ce.Message, ce.ExitCode)
	}
	if got := r.buildError(in, notFound); got != notFound {
		t.Errorf("driver flag read before another flag, not found: got %v, want the error unchanged", got)
	}
	in.primary = "y"
	if ce := r.buildError(in, notFound).(cmderrors.CommandError); ce.ExitCode != cmderrors.ExitError {
		t.Errorf("secondary flag with another driver value, not found: got %q exit=%d", ce.Error(), ce.ExitCode)
	}

	// any other error is left alone
	other := errors.New("boom")
	if got := r.buildError(in, other); got != other {
		t.Errorf("got %v, want the error unchanged", got)
	}
}
