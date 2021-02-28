package cmderrors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

const (
	// ErrTypeServer server error type. Server responded with an error
	ErrTypeServer = "serverError"

	// ErrTypeCommand command error. Local command error related to the usage of the tool etc.
	ErrTypeCommand = "commandError"
)

// CommandError is an error used to signal different error situations in command handling.
type CommandError struct {
	ErrorType       string `json:"errorType,omitempty"`
	Message         string `json:"message,omitempty"`
	silent          bool
	StatusCode      int                `json:"statusCode,omitempty"`
	ExitCode        int                `json:"exitCode,omitempty"`
	URL             string             `json:"url,omitempty"`
	CumulocityError *c8y.ErrorResponse `json:"c8yResponse,omitempty"`
	err             error
}

func (c CommandError) Error() string {
	details := ""
	if c.StatusCode > 0 {
		details = fmt.Sprintf(" ::StatusCode=%d", c.StatusCode)
	}
	if c.err != nil {
		details = fmt.Sprintf("%s", c.err)
	}
	return c.Message + details
}

// IsSilent returns true if the error should be silent
func (c CommandError) IsSilent() bool {
	return c.silent
}

// JSONString returns the json representation of the error
func (c CommandError) JSONString() string {
	out, err := json.Marshal(c)
	if err != nil {
		return fmt.Sprintf(`{"errorType":"%s", "message":"unexpected error. %s"}`, ErrTypeCommand, err)
	}
	return string(out)
}

// NewUserError creates a new user error
func NewUserError(a ...interface{}) CommandError {
	return CommandError{Message: fmt.Sprint(a...), ErrorType: ErrTypeCommand, ExitCode: 101, silent: false}
}

// NewUserErrorWithExitCode creates a user with a specific exit code
func NewUserErrorWithExitCode(exitCode int, a ...interface{}) CommandError {
	return CommandError{Message: fmt.Sprint(a...), ErrorType: ErrTypeCommand, ExitCode: exitCode, silent: false}
}

// NewSystemError creates a system error
func NewSystemError(a ...interface{}) CommandError {
	return CommandError{Message: fmt.Sprint(a...), ErrorType: ErrTypeCommand, ExitCode: 100, silent: false}
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

// NewServerError creates a server error from a Cumulocity response
func NewServerError(r *c8y.Response, err error) CommandError {
	cmdError := CommandError{
		Message:    err.Error(),
		ErrorType:  ErrTypeServer,
		silent:     false,
		ExitCode:   99,
		StatusCode: 0,
		err:        err,
	}

	if errors.Is(err, context.DeadlineExceeded) {
		cmdError.Message = "command timed out"
		cmdError.ExitCode = 106
	}

	if v, ok := err.(*c8y.ErrorResponse); ok {
		cmdError.CumulocityError = v
		cmdError.Message = v.Message
	}

	if r != nil {
		if r.Response != nil {
			cmdError.StatusCode = r.Response.StatusCode

			if r.Request != nil {
				cmdError.URL = r.Request.URL.Path
			}

			if v, ok := httpStatusCodeToExitCode[cmdError.StatusCode]; ok {
				cmdError.ExitCode = v
			}
		}
	}

	return cmdError
}

// NewSystemErrorF creates a custom system error
func NewSystemErrorF(format string, a ...interface{}) CommandError {
	return CommandError{Message: fmt.Sprintf(format, a...), ErrorType: ErrTypeCommand}
}

// NewErrorSummary create a error summary from a chanell of errors
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
