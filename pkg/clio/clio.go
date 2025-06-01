package clio

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

// isInputPiped checks if os.Stdin is receiving piped input.
// It does this by checking the file mode of os.Stdin.
// A named pipe (FIFO) typically indicates piped input.
// A character device indicates an interactive terminal.
// Any other mode might indicate a redirected file or other non-interactive source.
func IsInputPiped(f *os.File) bool {
	// return IsNamedPipe(f) || !IsInteractiveTerminal(f)
	return IsNamedPipe(f) && !IsInteractiveTerminal(f) // works ok, but does not handle empty input silently
}

func IsNamedPipe(f *os.File) bool {
	// Get file information for standard input.
	if f == nil {
		f = os.Stdin
	}
	fileInfo, err := f.Stat()
	if err != nil {
		// Return the error if we can't get stat info for stdin.
		return false
	}

	// Check if the mode indicates a named pipe (FIFO).
	// This is the most reliable way to detect piped input.
	return (fileInfo.Mode() & os.ModeNamedPipe) != 0
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

func IsWritable(f *os.File) bool {
	if f == nil {
		f = os.Stdin
	}

	// test if io is writable
	inputWritable := false
	if _, writeErr := f.Write([]byte("")); writeErr == nil {
		inputWritable = true
	}
	// fmt.Fprintf(os.Stderr, "Debug Stdin FD: inputWritable=%v\n", inputWritable)
	return inputWritable

	// stat, err := f.Stat()
	// if err != nil {
	// 	return false
	// }

	// mode := stat.Mode().Perm().String()
	// fmt.Fprintf(os.Stderr, "Debug Stdin FD: mode=%s\n", mode)
	// if len(mode) < 4 {
	// 	return false
	// }
	// return strings.Contains(mode[3:], "w")
}

func ConfigMaybePiped(f *os.File) bool {
	if f == nil {
		f = os.Stdin
	}

	stat, err := f.Stat()
	if err != nil {
		return false
	}

	isPipe := (stat.Mode()&os.ModeNamedPipe) != 0 ||
		(stat.Mode()&(os.ModeCharDevice|os.ModeDir|os.ModeSymlink)) == 0

	return isPipe
}

var EmptyMarker = []byte("\u0000")

func WriteEmptyMarker(w io.Writer) (int, error) {
	switch v := w.(type) {
	case *os.File:
		fmt.Fprintf(os.Stderr, "fName=%s\n", v.Name())
		if !strings.EqualFold(v.Name(), "stdout") {
			if IsNamedPipe(v) {
				return v.Write(EmptyMarker)
			}
		}

	default:
		if !IsTerminal(os.Stdout) {
			return v.Write(EmptyMarker)
		}
	}
	return 0, nil
}

func IsEmptyMarker(b []byte) bool {
	return bytes.EqualFold(b, EmptyMarker)
}

func IsTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
