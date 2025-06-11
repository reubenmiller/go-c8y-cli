package clio

import (
	"os"
	"runtime"
)

// HasPipedInput checks if os.Stdin is receiving piped input.
// It does this by checking the file mode of os.Stdin.
// A named pipe (FIFO) typically indicates piped input.
// A character device indicates an interactive terminal.
// Any other mode might indicate a redirected file or other non-interactive source.
func HasPipedInput(f *os.File) bool {
	// Get file information for standard input.
	if f == nil {
		f = os.Stdin
	}
	fileInfo, err := f.Stat()
	if err != nil {
		// Return the error if we can't get stat info for stdin.
		return false
	}

	if (fileInfo.Mode() & os.ModeNamedPipe) != 0 {
		return true
	} else if (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		// Stdin is a character device (e.g., a terminal)
		return false
	} else if fileInfo.Mode().IsRegular() {
		// Stdin is a regular file
		return true
	} else {
		// Stdin is something else (e.g., a socket)
		return true
	}
}

// IsInteractiveTerminal checks if os.Stdin is connected to an interactive terminal.
// This is true if the stdin is a character device.
func IsInteractiveTerminal(f *os.File) bool {
	if f == nil {
		f = os.Stdin
	}
	fileInfo, err := f.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func GetTTYStdin() *os.File {
	stdIn := os.Stdin
	if !IsInteractiveTerminal(stdIn) {
		// stdin is handling Piped input, so we have to prompt on a different input
		if runtime.GOOS == "windows" {
			if file, err := os.Open("CON"); err == nil {
				stdIn = file
			}
		} else {
			tty, err := os.Open("/dev/tty")
			if err == nil {
				stdIn = tty
			}
		}
	}
	return stdIn
}
