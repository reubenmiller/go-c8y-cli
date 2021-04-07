package cmderrors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

const (
	// ErrTypeServer server error type. Server responded with an error
	ErrTypeServer = "serverError"

	// ErrTypeCommand command error. Local command error related to the usage of the tool etc.
	ErrTypeCommand = "commandError"
)

type ExitCode int

const (
	ExitOK ExitCode = 0

	// Map HTTP status codes to exit codes
	ExitBadRequest400          ExitCode = 40
	ExitUnauthorized401        ExitCode = 1
	ExitForbidden403           ExitCode = 3
	ExitNotFound404            ExitCode = 4
	ExitMethodNotAllowed405    ExitCode = 5
	ExitConflict409            ExitCode = 9
	ExitExecutionTimeout413    ExitCode = 13
	ExitInvalidData422         ExitCode = 22
	ExitTooManyRequests429     ExitCode = 29
	ExitInternalServerError500 ExitCode = 50
	ExitNotImplemented501      ExitCode = 51
	ExitBadGateway502          ExitCode = 52
	ExitServiceUnavailable503  ExitCode = 53

	ExitGatewayTimeout504          ExitCode = 54
	ExitHTTPVersionNotSupported505 ExitCode = 55
	ExitVariantAlsoNegotiates506   ExitCode = 56
	ExitInsufficientStorage507     ExitCode = 57
	ExitLoopDetected508            ExitCode = 58

	ExitCancel              ExitCode = 2
	ExitError               ExitCode = 100
	ExitUserError           ExitCode = 101
	ExitNoSession           ExitCode = 102
	ExitAbortedWithErrors   ExitCode = 103
	ExitCompletedWithErrors ExitCode = 104
	ExitJobLimitExceeded    ExitCode = 105
	ExitTimeout             ExitCode = 106
	ExitInvalidAlias        ExitCode = 107
	ExitDecryption          ExitCode = 108
)

// CommandError is an error used to signal different error situations in command handling.
type CommandError struct {
	ErrorType       string `json:"errorType,omitempty"`
	Message         string `json:"message,omitempty"`
	silent          bool
	StatusCode      int                `json:"statusCode,omitempty"`
	ExitCode        ExitCode           `json:"exitCode,omitempty"`
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
	return strings.Join([]string{c.ErrorType + ":", c.Message, details}, " ")
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
	return CommandError{Message: fmt.Sprint(a...), ErrorType: ErrTypeCommand, ExitCode: ExitUserError, silent: false}
}

// NewUserErrorWithExitCode creates a user with a specific exit code
func NewUserErrorWithExitCode(exitCode ExitCode, a ...interface{}) CommandError {
	return CommandError{Message: fmt.Sprint(a...), ErrorType: ErrTypeCommand, ExitCode: exitCode, silent: false}
}

// NewSystemError creates a system error
func NewSystemError(a ...interface{}) CommandError {
	return CommandError{Message: fmt.Sprint(a...), ErrorType: ErrTypeCommand, ExitCode: ExitError, silent: false}
}

var httpStatusCodeToExitCode = map[int]ExitCode{
	400: ExitBadRequest400,
	401: ExitUnauthorized401,
	403: ExitForbidden403,
	404: ExitNotFound404,
	405: ExitMethodNotAllowed405,
	409: ExitConflict409,
	413: ExitExecutionTimeout413,
	422: ExitInvalidData422,
	429: ExitTooManyRequests429,
	500: ExitInternalServerError500,
	501: ExitNotImplemented501,
	502: ExitBadGateway502,
	503: ExitServiceUnavailable503,
	504: ExitGatewayTimeout504,
	505: ExitHTTPVersionNotSupported505,
	506: ExitVariantAlsoNegotiates506,
	507: ExitInsufficientStorage507,
	508: ExitLoopDetected508,
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
