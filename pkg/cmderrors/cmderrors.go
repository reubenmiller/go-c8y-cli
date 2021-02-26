package cmderrors

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

// CommandError is an error used to signal different error situations in command handling.
type CommandError struct {
	s           string
	userError   bool
	serverError bool
	silent      bool
	ExitCode    int
	statusCode  int
	c8yError    *c8y.ErrorResponse
	err         error
}

func (c CommandError) Error() string {
	details := ""
	if c.statusCode > 0 {
		details = fmt.Sprintf(" ::statusCode=%d", c.statusCode)
	}
	if c.err != nil {
		details = fmt.Sprintf("%s", c.err)
	}
	return c.s + details
}

func (c CommandError) StatusCode() int {
	return c.StatusCode()
}

func (c CommandError) IsUserError() bool {
	return c.userError
}

func (c CommandError) IsServerError() bool {
	return c.serverError
}

func (c CommandError) IsSilent() bool {
	return c.silent
}

func NewUserError(a ...interface{}) CommandError {
	return CommandError{s: fmt.Sprintln(a...), userError: true, ExitCode: 101, silent: false}
}

func NewUserErrorWithExitCode(exitCode int, a ...interface{}) CommandError {
	return CommandError{s: fmt.Sprintln(a...), userError: true, ExitCode: exitCode, silent: false}
}

func NewSystemError(a ...interface{}) CommandError {
	return CommandError{s: fmt.Sprintln(a...), userError: false, ExitCode: 100, silent: false}
}

var httpStatusCodeToExitCode = map[int]int{
	400: 40,
	401: 1,
	403: 3,
	404: 4,
	405: 5,
	409: 9,
	413: 13,
	422: 22,
	429: 29,
	500: 50,
	501: 51,
	502: 52,
	503: 53,
	504: 54,
	505: 55,
	506: 56,
	507: 57,
	508: 58,
}

func NewServerError(r *c8y.Response, err error) CommandError {
	message := ""
	exitCode := 99
	statusCode := 0
	var c8yError *c8y.ErrorResponse

	if errors.Is(err, context.DeadlineExceeded) {
		message = "command timed out"
		exitCode = 106
		err = nil
	}

	if v, ok := err.(*c8y.ErrorResponse); ok {
		c8yError = v
	}

	if r != nil {
		if r.Response != nil {
			statusCode = r.Response.StatusCode
			if v, ok := httpStatusCodeToExitCode[statusCode]; ok {
				exitCode = v
			}
		}
	}

	return CommandError{
		s:           message,
		userError:   false,
		serverError: true,
		silent:      false,
		ExitCode:    exitCode,
		statusCode:  statusCode,
		err:         err,
		c8yError:    c8yError,
	}
}

func NewSystemErrorF(format string, a ...interface{}) CommandError {
	return CommandError{s: fmt.Sprintf(format, a...), userError: false}
}

// Catch some of the obvious user errors from Cobra.
// We don't want to show the usage message for every error.
// The below may be to generic. Time will show.
var userErrorRegexp = regexp.MustCompile("argument|flag|shorthand")

func IsUserError(err error) bool {
	if cErr, ok := err.(CommandError); ok && cErr.IsUserError() {
		return true
	}

	return userErrorRegexp.MatchString(err.Error())
}

func IsServerError(err error) bool {
	if cErr, ok := err.(CommandError); ok && cErr.IsServerError() {
		return true
	}

	return false
}

func IsSilentError(err error) bool {
	if cErr, ok := err.(CommandError); ok && cErr.IsSilent() {
		return true
	}

	return false
}

func NewErrorSummary(message string, errorsCh <-chan error) error {
	errorSummary := errors.New(message)
	hasError := false
	for err := range errorsCh {
		if err != nil {
			errorSummary = errors.WithStack(err)
			hasError = true
		}
	}
	if !hasError {
		return nil
	}
	return errorSummary
}
