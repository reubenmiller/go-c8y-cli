package c8ystream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	neturl "net/url"
	"regexp"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
)

// toServerError turns a go-c8y v2 HTTP error into a cmderrors.CommandError that
// carries the HTTP status code and the mapped exit code, the way the v1 request
// handler did with NewServerError. Without this, the streaming path surfaced a
// raw core.Error which the final error handler treated as an unexpected error
// (always exit 100, status-code-blind), so --silentStatusCodes never matched
// and the exit code never reflected the actual HTTP status.
//
// Non-HTTP errors (resolution failures, validation, etc.) are returned
// unchanged so the generic handling still applies.
func (r *Runner) toServerError(err error) error {
	if err == nil {
		return nil
	}
	// A request that exceeded --timeout surfaces as a context deadline. Map it to
	// the dedicated timeout exit code (106) and the stable "command timed out"
	// message, matching v1's NewServerError. The sentinel error is used as Err so
	// it serialises to "error":{} rather than leaking the wrapped *url.Error.
	if errors.Is(err, context.DeadlineExceeded) {
		return cmderrors.CommandError{
			Message:   "command timed out",
			ErrorType: cmderrors.ErrTypeCommand,
			ExitCode:  cmderrors.ExitTimeout,
			Err:       context.DeadlineExceeded,
			IO:        r.Factory.IOStreams,
		}
	}
	var apiErr *core.Error
	if !errors.As(err, &apiErr) {
		return err
	}
	status := apiErr.StatusCode()
	if status < 100 {
		// Code is a sentinel (ErrorFromError/String/...), not an HTTP status.
		return err
	}

	// A name that the SDK could not resolve to an id is reported as v1 did:
	// "referenceByName: no matching items found" (commandError, exit code 4).
	if name, ok := unresolvedName(apiErr); ok {
		notFound := cmderrors.NewUserErrorWithExitCode(
			cmderrors.ExitNotFound404,
			cmderrors.NewNoMatchesFoundError(name),
		)
		notFound.StatusCode = http.StatusNotFound
		return notFound
	}

	detail := parseServerError(apiErr)

	// A failed request of an assert command (e.g. the lookup of the device)
	// is a plain command error ("METHOD url: status type message") as in v1.
	if r.assert != nil && detail != nil {
		return cmderrors.CommandError{
			Message:   detail.Error(),
			ErrorType: cmderrors.ErrTypeCommand,
			ExitCode:  cmderrors.ExitError,
		}
	}

	var body json.RawMessage
	if apiErr.MessageRaw != "" && json.Valid([]byte(apiErr.MessageRaw)) {
		body = json.RawMessage(apiErr.MessageRaw)
	}
	url := ""
	if apiErr.Response != nil && apiErr.Response.Request != nil {
		url = apiErr.Response.Request.URL.Path
	}

	// Looking up the bootstrap user of an application (microservices serviceusers
	// get) was a user error (exit code 101) in v1, whatever the status.
	if detail != nil && strings.HasSuffix(url, "/bootstrapUser") {
		return cmderrors.NewUserError(detail)
	}

	// The error is shaped like the v1 ErrorResponse: the message is the server's
	// message, the error carries "METHOD url: status type message" and, as JSON
	// (--withError), the {error, message, info} of the response. The raw body is
	// printed below the message unless --withError sends it to stdout.
	var wrapped error = apiErr
	if detail != nil {
		wrapped = detail
	}
	message := apiErr.Message
	if detail != nil {
		message = detail.Message
	}

	return cmderrors.CommandError{
		Message:         message,
		ErrorType:       cmderrors.ErrTypeServer,
		StatusCode:      status,
		ExitCode:        cmderrors.ExitCodeFromStatusCode(status),
		URL:             url,
		CumulocityError: body,
		Err:             wrapped,
		WithRawMessage:  !r.Config.WithError(),
		IO:              r.Factory.IOStreams,
	}
}

// WaitRequestError turns the HTTP error that ended a wait command (operations
// wait, inventory wait) into the command error v1 gave: "METHOD url: status
// type message", commandError, exit code 100. Other errors are returned as is.
func WaitRequestError(err error) error {
	var apiErr *core.Error
	if err == nil || !errors.As(err, &apiErr) || apiErr.StatusCode() < 100 {
		return err
	}
	detail := parseServerError(apiErr)
	if detail == nil {
		return err
	}
	return cmderrors.CommandError{
		Message:   detail.Error(),
		ErrorType: cmderrors.ErrTypeCommand,
		ExitCode:  cmderrors.ExitError,
	}
}

// ServerResponseError is the error of a failed request in the format of the
// v1 client: it prints as "METHOD url: status type message" and marshals to
// the error body of the server ({error, message, info}).
type ServerResponseError struct {
	Method     string `json:"-"`
	URL        string `json:"-"`
	StatusCode int    `json:"-"`
	ErrorType  string `json:"error,omitempty"`
	Message    string `json:"message,omitempty"`
	Info       string `json:"info,omitempty"`

	err error
}

func (e *ServerResponseError) Error() string {
	return fmt.Sprintf("%s %s: %d %s %s", e.Method, e.URL, e.StatusCode, e.ErrorType, e.Message)
}

// Unwrap gives access to the SDK error (status code checks, errors.As).
func (e *ServerResponseError) Unwrap() error { return e.err }

// parseServerError reads the error body of the server. It returns nil when
// the request is not known (no response) so the caller keeps the SDK error.
func parseServerError(apiErr *core.Error) *ServerResponseError {
	if apiErr.Response == nil || apiErr.Response.Request == nil || apiErr.Response.Request.URL == nil {
		return nil
	}
	e := &ServerResponseError{
		Method:     apiErr.Response.Request.Method,
		URL:        sanitizeURL(apiErr.Response.Request.URL),
		StatusCode: apiErr.StatusCode(),
		err:        apiErr,
	}
	var body struct {
		Error    string   `json:"error"`
		Message  string   `json:"message"`
		Messages []string `json:"messages"`
		Info     string   `json:"info"`
	}
	if apiErr.MessageRaw != "" && json.Unmarshal([]byte(apiErr.MessageRaw), &body) == nil {
		e.ErrorType = body.Error
		e.Message = body.Message
		if e.Message == "" {
			e.Message = strings.Join(body.Messages, "; ")
		}
		e.Info = body.Info
		if e.Message != "" || e.ErrorType != "" {
			return e
		}
	}
	// no usable error body (e.g. an HTML page of a proxy): keep the text the SDK derived
	e.ErrorType, e.Info = "", ""
	e.Message = apiErr.Message
	return e
}

// sanitizeURL is the url of a request without the secrets in its query.
func sanitizeURL(u *neturl.URL) string {
	q := u.Query()
	if q.Get("client_secret") == "" {
		return u.String()
	}
	q.Set("client_secret", "REDACTED")
	clean := *u
	clean.RawQuery = q.Encode()
	return clean.String()
}

var nameLookupRe = regexp.MustCompile(`not found with name[:=]\s*(.+?)(?:,\s*\w+[:=].*)?$`)

// unresolvedName returns the name of a lookup by name that found nothing: the
// SDK's "<thing> not found with name: <name>", or the 404 of the group name
// endpoint. A 404 of an id is a normal server error.
func unresolvedName(apiErr *core.Error) (string, bool) {
	if apiErr.StatusCode() != http.StatusNotFound {
		return "", false
	}
	if apiErr.Response == nil {
		if m := nameLookupRe.FindStringSubmatch(apiErr.Message); m != nil {
			return m[1], true
		}
		return "", false
	}
	if req := apiErr.Response.Request; req != nil && req.URL != nil {
		if _, name, ok := strings.Cut(req.URL.Path, "/groupByName/"); ok && name != "" {
			return name, true
		}
	}
	return "", false
}

// errorDocument converts an item error into the JSON document emitted to stdout
// under --withError: the status-aware CommandError (errorType, message,
// statusCode, c8yResponse) marshalled to JSON, so the output template and select
// stages can shape it like any other document. Non-CommandError errors (which
// have no structured body) are not emitted.
//
// When an output template is active it is applied here: success documents are
// templated per job by the producer wrapper, which error documents bypass (the
// collector emits them downstream), so without this the error would reach the
// output unshaped.
func (r *Runner) errorDocument(err error) (jsondoc.JSONDoc, bool) {
	var ce cmderrors.CommandError
	if !errors.As(r.toServerError(err), &ce) {
		return jsondoc.Empty(), false
	}
	raw, marshalErr := json.Marshal(ce)
	if marshalErr != nil {
		return jsondoc.Empty(), false
	}
	if r.tmpl != nil {
		var input any
		if len(r.tmplInput) > 0 {
			input = r.tmplInput
		}
		r.tmplMu.Lock()
		reqData, respData := r.tmplRequest, r.tmplResponse
		r.tmplMu.Unlock()
		out, tmplErr := r.tmpl.Evaluate(raw, input, reqData, respData, r.tmplFlags)
		if tmplErr != nil {
			return jsondoc.Empty(), false
		}
		return jsondoc.New(unwrapTemplateString(out)), true
	}
	return jsondoc.New(raw), true
}

// collapseErrors reduces the collected item errors to the single command result
// error, applying toServerError so the status code/exit code/body are present
// for the downstream silent-status-code and exit-code handling. A batch with
// more than one failure reports only the aggregate ("jobs completed with N
// errors", exit 104) — the individual errors were already logged as they
// occurred (ErrorCollector.Log), matching the v1 worker output shape.
func (r *Runner) collapseErrors(errs []error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return r.toServerError(errs[0])
	default:
		aggregate := cmderrors.NewErrorWithExitCode(
			cmderrors.ExitCompletedWithErrors,
			nil,
			fmt.Sprintf("jobs completed with %d errors", len(errs)),
		)
		// The assert commands of v1 did not print a summary line: the
		// assertion errors were all reported, and the exit code says the rest.
		aggregate.Processed = r.assert != nil
		return aggregate
	}
}
