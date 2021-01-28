Enum c8yExitCode {
    StatusBadRequest400 = 40
	StatusUnauthorized401 = 1
	StatusForbidden403 = 3
	StatusNotFound404 = 4
	StatusMethodNotAllowed405 = 5
	StatusConflict409 = 9
	StatusExecutionTimeout413 = 13
	StatusInvalidData422 = 22
	StatusTooManyRequests429 = 29
	StatusInternalServerError500 = 50
	StatusNotImplemented501 = 51
	StatusBadGateway502 = 52
	StatusServiceUnavailable503 = 53
	
	# Command errors
    SystemError = 100
    UserError = 101
    NoSessionLoaded = 102
    BatchAbortedWithErrors = 103
    BatchCompletedWithErrors = 104
}