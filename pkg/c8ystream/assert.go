package c8ystream

import (
	"errors"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
)

// Assert makes the command an assertion. An item whose assertion fails
// (an error wrapping cmderrors.ErrAssertion) is dropped silently unless strict
// is set, so the passing items can be piped on. Any error the command ends
// with gets the assertion exit code.
func (r *Runner) Assert(strict bool) {
	r.assert = &assertion{strict: strict}
}

type assertion struct {
	strict bool
}

// filter drops failed assertions in non-strict mode.
func (a *assertion) filter(seq func(yield func(jsondoc.JSONDoc, error) bool)) func(yield func(jsondoc.JSONDoc, error) bool) {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		for doc, err := range seq {
			if err != nil && !a.strict && errors.Is(err, cmderrors.ErrAssertion) {
				continue
			}
			if !yield(doc, err) {
				return
			}
		}
	}
}

// exitCode gives a processed command error the assertion exit code. The
// errors of the batch itself (aborted, completed with errors, job limit) keep
// their own exit code: the individual assertion errors were already given the
// assertion exit code when they were reported, as in v1.
func (a *assertion) exitCode(err error) error {
	if err == nil {
		return nil
	}
	var ce cmderrors.CommandError
	if errors.As(err, &ce) {
		switch ce.ExitCode {
		case cmderrors.ExitAbortedWithErrors, cmderrors.ExitCompletedWithErrors, cmderrors.ExitJobLimitExceeded:
			return err
		}
	}
	wrapped := cmderrors.NewUserErrorWithExitCode(cmderrors.ExitAssertionError, err)
	wrapped.Processed = true
	return wrapped
}
