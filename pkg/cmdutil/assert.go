package cmdutil

import (
	"errors"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
)

func (f *Factory) WithExitCodeError(exitCode cmderrors.ExitCode) func(error) error {
	return func(err error) error {
		err = f.CheckPostCommandError(err)
		if err == nil {
			return nil
		}
		wrappedErr := cmderrors.NewUserErrorWithExitCode(exitCode, err)
		wrappedErr.Processed = true
		return wrappedErr
	}
}

func ProcessAssertError(f *Factory, strictMode bool) func(error) error {
	return func(err error) error {
		if !strictMode && errors.Is(err, cmderrors.ErrAssertion) {
			// ignore assertion errors
			return nil
		}
		return f.WithExitCodeError(cmderrors.ExitAssertionError)(err)
	}
}
