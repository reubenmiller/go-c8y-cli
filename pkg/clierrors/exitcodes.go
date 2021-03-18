package clierrors

type ExitCode int

const (
	ExitOK     ExitCode = 0
	ExitError  ExitCode = 1
	ExitCancel ExitCode = 2
	ExitAuth   ExitCode = 4
)
