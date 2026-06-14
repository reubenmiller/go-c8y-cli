package c8ystream

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
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
	var apiErr *core.Error
	if !errors.As(err, &apiErr) {
		return err
	}
	status := apiErr.StatusCode()
	if status < 100 {
		// Code is a sentinel (ErrorFromError/String/...), not an HTTP status.
		return err
	}

	var body json.RawMessage
	if apiErr.MessageRaw != "" && json.Valid([]byte(apiErr.MessageRaw)) {
		body = json.RawMessage(apiErr.MessageRaw)
	}
	url := ""
	if apiErr.Response != nil && apiErr.Response.Request != nil {
		url = apiErr.Response.Request.URL.Path
	}

	return cmderrors.CommandError{
		Message:         apiErr.Message,
		ErrorType:       cmderrors.ErrTypeServer,
		StatusCode:      status,
		ExitCode:        cmderrors.ExitCodeFromStatusCode(status),
		URL:             url,
		CumulocityError: body,
		Err:             err,
		WithRawMessage:  r.Config.WithError(),
		IO:              r.Factory.IOStreams,
	}
}

// collapseErrors reduces the collected item errors to the single command result
// error, applying toServerError so the status code/exit code/body are present
// for the downstream silent-status-code and exit-code handling. A batch with
// more than one failure reports ExitCompletedWithErrors, carrying the first
// (converted) error for context — matching the v1 worker aggregate.
func (r *Runner) collapseErrors(errs []error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return r.toServerError(errs[0])
	default:
		first := r.toServerError(errs[0])
		return cmderrors.NewErrorWithExitCode(
			cmderrors.ExitCompletedWithErrors,
			first,
			fmt.Sprintf("jobs completed with %d errors", len(errs)),
		)
	}
}
