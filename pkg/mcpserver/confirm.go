package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// DefaultConfirmTimeout how long the user has to confirm a device command
const DefaultConfirmTimeout = 2 * time.Minute

// maxDialogText limit for the command shown in a dialog
const maxDialogText = 1500

var errNoDialog = errors.New("no dialog available")

// DeviceShellConfirmer asks the user to confirm commands which run on a device. It uses
// a native dialog if available, so the user doesn't have to switch to the server's
// terminal, and falls back to the terminal otherwise. The MCP client can't answer it.
type DeviceShellConfirmer struct {
	// Stdin, Stderr server terminal
	Stdin  *os.File
	Stderr io.Writer

	// Terminal lock which is held while the terminal is used
	Terminal sync.Locker

	// Notify optional callback to draw the user's attention to the terminal prompt
	Notify func(message string)

	// Timeout how long the user has to answer. Defaults to DefaultConfirmTimeout
	Timeout time.Duration

	// NoDialog only use the terminal
	NoDialog bool

	// dialog shows the native dialog (replaced in tests)
	dialog func(ctx context.Context, title string, message string, timeout time.Duration) (bool, error)
}

// Confirm returns nil if the user allows the command
func (c *DeviceShellConfirmer) Confirm(ctx context.Context, session SessionInfo, command string) error {
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = DefaultConfirmTimeout
	}
	target := session.Host
	if session.Tenant != "" {
		target += " (tenant " + session.Tenant + ")"
	}
	message := fmt.Sprintf("The agent wants to run a command on a device of %s:\n\n%s", target, truncate(command, maxDialogText))

	if !c.NoDialog {
		dialog := c.dialog
		if dialog == nil {
			dialog = nativeDialog
		}
		allowed, err := dialog(ctx, "c8y mcp serve", message, timeout)
		if err == nil {
			if !allowed {
				return errors.New("the user did not allow running the command on the device. Don't retry it unless the user asks you to")
			}
			return nil
		}
		if !errors.Is(err, errNoDialog) {
			fmt.Fprintf(c.Stderr, "Could not show a dialog, asking in the terminal instead. %s\n", err)
		}
	}
	return c.confirmInTerminal(ctx, message, timeout)
}

func (c *DeviceShellConfirmer) confirmInTerminal(ctx context.Context, message string, timeout time.Duration) error {
	if c.Stdin == nil {
		return errors.New("the command could not be confirmed, as no dialog is available and the server has no terminal")
	}
	if c.Terminal != nil {
		c.Terminal.Lock()
		defer c.Terminal.Unlock()
	}
	fmt.Fprintf(c.Stderr, "\n%s\n\nAllow this command? [y/N] \a", message)
	if c.Notify != nil {
		go c.Notify("The agent wants to run a command on a device. Confirm it in the c8y mcp serve terminal.")
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if _, err := waitReadable(c.Stdin, 0); err != nil {
		return fmt.Errorf("the command could not be confirmed in the terminal. %w", err)
	}
	for {
		if ctx.Err() != nil {
			fmt.Fprintln(c.Stderr)
			return errors.New("the user did not confirm running the command on the device in time")
		}
		line, ok, err := readLineIfReady(c.Stdin, pollInterval)
		if ok {
			switch strings.ToLower(strings.TrimSpace(line)) {
			case "y", "yes":
				return nil
			}
			fmt.Fprintln(c.Stderr, "Command rejected")
			return errors.New("the user did not allow running the command on the device. Don't retry it unless the user asks you to")
		}
		if err != nil {
			return fmt.Errorf("the command could not be confirmed in the terminal. %w", err)
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// nativeDialog asks the user with a native dialog. The values are passed as arguments,
// so they are never interpreted as a script. Deny is the default button.
func nativeDialog(ctx context.Context, title string, message string, timeout time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout+5*time.Second)
	defer cancel()

	switch runtime.GOOS {
	case "darwin":
		if os.Getenv("SSH_CONNECTION") != "" {
			// the dialog would be shown on the machine's screen, not to the remote user
			return false, errNoDialog
		}
		seconds := fmt.Sprintf("%d", int(timeout.Seconds()))
		out, err := exec.CommandContext(ctx, "osascript",
			"-e", "on run argv",
			"-e", "set r to display dialog (item 2 of argv) with title (item 1 of argv) buttons {\"Deny\", \"Allow\"} default button \"Deny\" cancel button \"Deny\" with icon caution giving up after ((item 3 of argv) as integer)",
			"-e", "if gave up of r then return \"timeout\"",
			"-e", "return button returned of r",
			"-e", "end run",
			title, message, seconds,
		).Output()
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && strings.Contains(string(exitErr.Stderr), "User canceled") {
				return false, nil
			}
			return false, err
		}
		return strings.TrimSpace(string(out)) == "Allow", nil

	case "linux", "freebsd", "openbsd", "netbsd":
		if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
			return false, errNoDialog
		}
		if path, err := exec.LookPath("zenity"); err == nil {
			err := exec.CommandContext(ctx, path, "--question", "--default-cancel", "--title", title, "--text", message,
				"--ok-label", "Allow", "--cancel-label", "Deny", "--timeout", fmt.Sprintf("%d", int(timeout.Seconds()))).Run()
			return dialogExitStatus(err)
		}
		if path, err := exec.LookPath("kdialog"); err == nil {
			err := exec.CommandContext(ctx, path, "--title", title, "--warningyesno", message, "--yes-label", "Allow", "--no-label", "Deny").Run()
			return dialogExitStatus(err)
		}
	}
	return false, errNoDialog
}

// dialogExitStatus exit code 0 means allowed; other exit codes mean denied or timed out
func dialogExitStatus(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false, nil
	}
	return false, err
}
