package clierrors

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

	ExitCancel              ExitCode = 2
	ExitError               ExitCode = 100
	ExitUserError           ExitCode = 101
	ExitNoSession           ExitCode = 102
	ExitAbortedWithErrors   ExitCode = 103
	ExitCompletedWithErrors ExitCode = 104
	ExitJobLimitExceeded    ExitCode = 105
	ExitTimeout             ExitCode = 106
)
