package mcpserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// SelectSession lets the user select a session in the terminal. The selection runs
// in a separate c8y process, as the interactive prompt can leave a pending read on
// the terminal, which would take the input meant for later prompts.
func SelectSession(ctx context.Context, executable string, stdin io.Reader, stderr io.Writer) (string, error) {
	cmd := exec.CommandContext(ctx, executable, "sessions", "list")
	cmd.Stdin = stdin
	cmd.Stderr = stderr
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("no session was selected. %w", err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// RunConsole reads commands which the user types in the server's terminal, while the
// terminal isn't used by a prompt or the login. Each command is a line of input, which
// is passed to handle. Input is only read once a whole line is available, so no input
// meant for other prompts is taken.
//
// It returns an error if the terminal can't be read without blocking (e.g. on Windows),
// or the terminal is closed.
func RunConsole(ctx context.Context, stdin *os.File, terminal sync.Locker, handle func(command string)) error {
	if _, err := waitReadable(stdin, 0); err != nil {
		return err
	}
	for ctx.Err() == nil {
		terminal.Lock()
		line, ok, err := readLineIfReady(stdin, pollInterval)
		terminal.Unlock()
		if ok {
			handle(strings.TrimSpace(line))
		}
		if err != nil {
			return err
		}
		if !ok {
			// give prompts which are waiting for the terminal a chance to take it
			time.Sleep(10 * time.Millisecond)
		}
	}
	return nil
}
