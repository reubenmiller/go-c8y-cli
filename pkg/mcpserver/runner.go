package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

// ciEnvVars environment variables which would activate the CI mode in c8y
// (which enables all commands), so they are removed from the child processes
var ciEnvVars = []string{
	"CI",
	"BUILD_ID",
	"BUILD_NUMBER",
	"RUN_ID",
	"CI_APP_ID",
	"CI_BUILD_ID",
	"CI_BUILD_NUMBER",
	"CI_NAME",
	"CONTINUOUS_INTEGRATION",
}

// controlledEnvPrefixes environment variables which control settings the MCP
// server is responsible for
var controlledEnvPrefixes = []string{
	"C8Y_MODE=",
	"C8Y_SETTINGS_CI=",
	"C8Y_SETTINGS_MODE",
	"C8Y_SETTINGS_SESSION_MODE=",
	"C8Y_SETTINGS_SESSION_CONFIRMATION=",
	"C8Y_SETTINGS_DEFAULTS_FORCE",
	"C8Y_SETTINGS_DEFAULTS_DRY",
	"C8Y_SETTINGS_LOGGER_HIDESENSITIVE=",
	"C8Y_SETTINGS_DEFAULTS_INSECURE=",
}

// Runner executes c8y commands as child processes
type Runner struct {
	// Executable path to the c8y binary
	Executable string

	// SessionFile session file which is used for all commands
	SessionFile string

	// Mode the maximum mode used for write commands
	Mode config.SessionMode

	// BaseEnv base environment for the child processes (defaults to os.Environ())
	BaseEnv []string

	// Auth optional credentials provider. When set, credentials inherited from
	// the server's environment are removed and the provider's credentials are
	// used instead
	Auth AuthProvider
}

// credentialEnvVars secrets which are passed to the commands via a pipe (see passCredentials)
var credentialEnvVars = map[string]func(c *config.Credentials, v string){
	"C8Y_TOKEN":    func(c *config.Credentials, v string) { c.Token = v },
	"C8Y_PASSWORD": func(c *config.Credentials, v string) { c.Password = v },
}

// splitCredentials removes the secrets from the environment variables
func splitCredentials(env []string) ([]string, *config.Credentials) {
	if env == nil {
		return nil, nil
	}
	out := make([]string, 0, len(env))
	var credentials *config.Credentials
	for _, item := range env {
		key, value, _ := strings.Cut(item, "=")
		if set, ok := credentialEnvVars[key]; ok {
			if credentials == nil {
				credentials = &config.Credentials{}
			}
			set(credentials, value)
			continue
		}
		out = append(out, item)
	}
	return out, credentials
}

// passCredentials passes the credentials to the command via an inherited pipe (read by
// config.ReadCredentialsFromFD), so they don't appear in the command's environment.
// Windows doesn't support inheriting extra file descriptors, so the environment is used.
func passCredentials(cmd *exec.Cmd, credentials *config.Credentials) error {
	if credentials == nil {
		return nil
	}
	if runtime.GOOS == "windows" {
		if credentials.Token != "" {
			cmd.Env = append(cmd.Env, "C8Y_TOKEN="+credentials.Token)
		}
		if credentials.Password != "" {
			cmd.Env = append(cmd.Env, "C8Y_PASSWORD="+credentials.Password)
		}
		if credentials.Passphrase != "" {
			cmd.Env = append(cmd.Env, "C8Y_PASSPHRASE="+credentials.Passphrase)
		}
		return nil
	}

	data, err := json.Marshal(credentials)
	if err != nil {
		return err
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		return err
	}
	// The credentials are small, so they fit into the pipe buffer before the command starts
	_, writeErr := writer.Write(data)
	closeErr := writer.Close()
	if writeErr != nil || closeErr != nil {
		reader.Close()
		return fmt.Errorf("could not pass the credentials to the command. %v %v", writeErr, closeErr)
	}
	// ExtraFiles[0] is file descriptor 3 in the command
	cmd.ExtraFiles = append(cmd.ExtraFiles, reader)
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%d", config.EnvCredentialsFD, 2+len(cmd.ExtraFiles)))
	return nil
}

// injectFlags adds flags to the arguments, before the argument terminator "--", as
// the arguments after it are passed to another program (e.g. the command run on the
// device via ssh)
func injectFlags(args []string, flags ...string) []string {
	before, after, found := SplitAtTerminator(args)
	out := make([]string, 0, len(args)+len(flags))
	out = append(out, before...)
	out = append(out, flags...)
	if found {
		out = append(out, "--")
		out = append(out, after...)
	}
	return out
}

// Result output of the executed command(s)
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	// Truncated is true if the stdout was truncated
	Truncated bool
}

// Env builds the environment for a child process
func (r *Runner) Env(mode config.SessionMode, authEnv []string) []string {
	base := r.BaseEnv
	if base == nil {
		base = os.Environ()
	}
	env := make([]string, 0, len(base)+len(authEnv)+6)
	for _, item := range base {
		if isControlledEnv(item) {
			continue
		}
		if authEnv != nil && isInheritedAuthEnv(item) {
			continue
		}
		env = append(env, item)
	}
	env = append(env,
		config.EnvSessionMode+"="+mode.String(),
		"C8Y_SETTINGS_CI=false",
		"C8Y_SETTINGS_LOGGER_HIDESENSITIVE=true",
	)
	if authEnv != nil {
		env = append(env, authEnv...)
	} else if r.SessionFile != "" {
		env = append(env, "C8Y_SESSION="+r.SessionFile)
	}
	return env
}

func isInheritedAuthEnv(item string) bool {
	name, _, _ := strings.Cut(item, "=")
	for _, v := range inheritedAuthEnvVars {
		if strings.EqualFold(name, v) {
			return true
		}
	}
	return false
}

func isControlledEnv(item string) bool {
	name, _, _ := strings.Cut(item, "=")
	for _, v := range ciEnvVars {
		if strings.EqualFold(name, v) {
			return true
		}
	}
	upper := strings.ToUpper(item)
	for _, prefix := range controlledEnvPrefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

// Run executes the plan. Read-only commands are executed in prod (read-only)
// mode so the existing c8y mode checks reject any write request, even if the
// command was misclassified.
func (r *Runner) Run(ctx context.Context, plan *Plan, stdin string, tag string, maxOutput int) (*Result, error) {
	var authEnv []string
	if r.Auth != nil {
		var err error
		authEnv, err = r.Auth.AuthEnv(ctx)
		if err != nil {
			return nil, err
		}
		if authEnv == nil {
			authEnv = []string{}
		}
	}
	// Pass the secrets via a pipe instead of the environment, which other processes
	// of the same user can read
	authEnv, credentials := splitCredentials(authEnv)
	readEnv := r.Env(config.SessionModeProduction, authEnv)
	writeEnv := r.Env(r.Mode, authEnv)

	cmds := make([]*exec.Cmd, 0, len(plan.Segments))
	stderrs := make([]*bytes.Buffer, 0, len(plan.Segments))
	for _, seg := range plan.Segments {
		injected := []string{"--logMessage", tag}
		env := readEnv
		if !seg.Read {
			// approval has already been handled by the MCP server
			injected = append(injected, "--force")
			env = writeEnv
		}
		cmd := exec.CommandContext(ctx, r.Executable, injectFlags(seg.Args, injected...)...)
		cmd.Env = env
		if err := passCredentials(cmd, credentials); err != nil {
			return nil, err
		}
		// Stop the whole process tree on timeout, e.g. the ssh process started by
		// "remoteaccess connect ssh", so it can't keep the output open
		setProcessGroup(cmd)
		cmd.WaitDelay = 5 * time.Second
		stderr := &bytes.Buffer{}
		cmd.Stderr = &limitedWriter{w: stderr, n: 20_000}
		cmds = append(cmds, cmd)
		stderrs = append(stderrs, stderr)
	}

	if stdin != "" {
		cmds[0].Stdin = strings.NewReader(stdin)
	}
	for i := 1; i < len(cmds); i++ {
		pipe, err := cmds[i-1].StdoutPipe()
		if err != nil {
			return nil, err
		}
		cmds[i].Stdin = pipe
	}

	stdout := &limitedWriter{w: &bytes.Buffer{}, n: maxOutput}
	cmds[len(cmds)-1].Stdout = stdout

	// The commands have their own copy of the credentials pipe once started
	defer func() {
		for _, cmd := range cmds {
			for _, f := range cmd.ExtraFiles {
				f.Close()
			}
		}
	}()

	for i, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			for _, started := range cmds[:i] {
				_ = started.Process.Kill()
			}
			return nil, fmt.Errorf("could not start c8y. %w", err)
		}
	}

	exitCode := 0
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				if exitCode == 0 {
					exitCode = exitErr.ExitCode()
				}
			} else if exitCode == 0 {
				exitCode = -1
			}
		}
	}

	stderrText := []string{}
	for i, buf := range stderrs {
		if buf.Len() == 0 {
			continue
		}
		prefix := ""
		if len(stderrs) > 1 {
			prefix = fmt.Sprintf("[%s] ", plan.Segments[i].CommandPath)
		}
		stderrText = append(stderrText, prefix+strings.TrimSpace(buf.String()))
	}

	result := &Result{
		Stdout:    stdout.w.(*bytes.Buffer).String(),
		Stderr:    strings.Join(stderrText, "\n"),
		ExitCode:  exitCode,
		Truncated: stdout.truncated,
	}
	if ctx.Err() != nil {
		return result, fmt.Errorf("command timed out or was cancelled. %w", ctx.Err())
	}
	return result, nil
}

// limitedWriter writes up to n bytes and silently discards the rest so the
// child process is not blocked
type limitedWriter struct {
	mu        sync.Mutex
	w         io.Writer
	n         int
	written   int
	truncated bool
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	remaining := l.n - l.written
	if remaining <= 0 {
		if len(p) > 0 {
			l.truncated = true
		}
		return len(p), nil
	}
	chunk := p
	if len(chunk) > remaining {
		chunk = chunk[:remaining]
		l.truncated = true
	}
	n, err := l.w.Write(chunk)
	l.written += n
	if err != nil {
		return n, err
	}
	return len(p), nil
}
