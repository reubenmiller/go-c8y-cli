package mcpserver

import (
	"context"
	"os/exec"
	"runtime"
	"time"
)

// DesktopNotification shows a desktop notification, if supported (best effort).
// The values are passed as arguments, so they are never interpreted as a script.
func DesktopNotification(title string, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	switch runtime.GOOS {
	case "darwin":
		_ = exec.CommandContext(ctx, "osascript",
			"-e", "on run argv",
			"-e", "display notification (item 2 of argv) with title (item 1 of argv) sound name \"Ping\"",
			"-e", "end run",
			title, message,
		).Run()
	case "linux", "freebsd", "openbsd", "netbsd":
		if path, err := exec.LookPath("notify-send"); err == nil {
			_ = exec.CommandContext(ctx, path, "--app-name=c8y", title, message).Run()
		}
	}
}
