package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

const (
	ansiEsc    = "\u001B"
	clearLine  = "\r\033[K"
	hideCursor = ansiEsc + "[?25l"
	showCursor = ansiEsc + "[?25h"
)

type cmder interface {
	getCommand() *cobra.Command
}

// commandError is an error used to signal different error situations in command handling.
type commandError struct {
	s           string
	userError   bool
	serverError bool
	silent      bool
	exitCode    int
	statusCode  int
}

func (c commandError) Error() string {
	details := ""
	if c.statusCode > 0 {
		details = fmt.Sprintf(" ::statusCode=%d", c.statusCode)
	}
	return c.s + details
}

func (c commandError) isUserError() bool {
	return c.userError
}

func (c commandError) isServerError() bool {
	return c.serverError
}

func (c commandError) isSilent() bool {
	return c.silent
}

func newUserError(a ...interface{}) commandError {
	return commandError{s: fmt.Sprintln(a...), userError: true, exitCode: 101}
}

func newUserErrorWithExitCode(exitCode int, a ...interface{}) commandError {
	return commandError{s: fmt.Sprintln(a...), userError: true, exitCode: exitCode}
}

func newSystemError(a ...interface{}) commandError {
	return commandError{s: fmt.Sprintln(a...), userError: false, exitCode: 100}
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

func newServerError(r *c8y.Response, err error) commandError {
	message := ""
	exitCode := 99
	statusCode := 0
	if r != nil {
		if r.Response != nil {
			statusCode = r.Response.StatusCode
			if v, ok := httpStatusCodeToExitCode[statusCode]; ok {
				exitCode = v
			}
		}
	}
	return commandError{
		s:           message,
		userError:   false,
		serverError: true,
		silent:      true,
		exitCode:    exitCode,
		statusCode:  statusCode,
	}
}

func newSystemErrorF(format string, a ...interface{}) commandError {
	return commandError{s: fmt.Sprintf(format, a...), userError: false}
}

// Catch some of the obvious user errors from Cobra.
// We don't want to show the usage message for every error.
// The below may be to generic. Time will show.
var userErrorRegexp = regexp.MustCompile("argument|flag|shorthand")

func isUserError(err error) bool {
	if cErr, ok := err.(commandError); ok && cErr.isUserError() {
		return true
	}

	return userErrorRegexp.MatchString(err.Error())
}

func isServerError(err error) bool {
	if cErr, ok := err.(commandError); ok && cErr.isServerError() {
		return true
	}

	return false
}

func isSilentError(err error) bool {
	if cErr, ok := err.(commandError); ok && cErr.isSilent() {
		return true
	}

	return false
}

func newErrorSummary(message string, errorsCh <-chan error) error {
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

// replacePathParameters replaces all of the path parameters in a given URI with the provided values
// Example:
// 	"alarm/alarms/{id}" => "alarm/alarms/1234" if given a parameter map of {"id": "1234"}
func replacePathParameters(uri string, parameters map[string]string) string {
	if parameters == nil {
		return uri
	}
	for key, value := range parameters {
		uri = strings.ReplaceAll(uri, "{"+key+"}", value)
	}
	return uri
}
