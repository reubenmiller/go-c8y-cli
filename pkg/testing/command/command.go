package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/shlex"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/root"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fakestdin"
)

func NewMockCommand() *root.CmdRoot {
	rootCmd, err := root.NewCommand("", "")
	if err != nil {
		panic(err)
	}
	return rootCmd
}

func ExecuteCmd(cmd *root.CmdRoot, cmdArgs interface{}, opts ...ExecuteOptions) error {
	removeCommandName := func(args []string) []string {
		if len(args) > 0 && args[0] == "c8y" {
			args = args[1:]
		}
		return args
	}
	args := make([]string, 0)
	switch v := cmdArgs.(type) {
	case string:
		if parsedArgs, err := shlex.Split(strings.TrimSpace(v)); err == nil {
			args = append(args, parsedArgs...)
		} else {
			args = append(args, strings.Split(strings.TrimSpace(v), " ")...)
		}
	case []string:
		args = append(args, v...)
	}

	cmd.SetArgs(removeCommandName(args))

	for _, opt := range opts {
		if err := opt(cmd); err != nil {
			return err
		}
	}

	return cmd.Execute()
}

func ExecuteCmdWithStandardOutput(cmd *root.CmdRoot, cmdArgs interface{}, opts ...ExecuteOptions) (stdout string, err error) {
	stdoutW := strings.Builder{}
	opts = append(opts, WithCaptureStdOut(&stdoutW))
	cmdErr := ExecuteCmd(
		cmd,
		cmdArgs,
		opts...,
	)
	return stdoutW.String(), cmdErr
}

func ExecuteCmdWithOutputs(cmd *root.CmdRoot, cmdArgs interface{}, opts ...ExecuteOptions) (stdout string, stderr string, err error) {
	stdoutW := strings.Builder{}
	stderrW := strings.Builder{}
	opts = append(opts, WithCaptureStdOut(&stdoutW), WithCaptureStdErr(&stderrW))
	cmdErr := ExecuteCmd(
		cmd,
		cmdArgs,
		opts...,
	)
	return stdoutW.String(), stderrW.String(), cmdErr
}

type ExecuteOptions func(*root.CmdRoot) error

func WithEmptyEnv(t *testing.T) ExecuteOptions {
	return WithEnv(t, map[string]string{})
}

func WithStdinTTY(v bool) ExecuteOptions {
	return func(cmd *root.CmdRoot) error {
		cmd.Factory.IOStreams.SetStdinTTY(v)
		return nil
	}
}

func WithStdoutTTY(v bool) ExecuteOptions {
	return func(cmd *root.CmdRoot) error {
		cmd.Factory.IOStreams.SetStdoutTTY(v)
		return nil
	}
}

func WithStderrTTY(v bool) ExecuteOptions {
	return func(cmd *root.CmdRoot) error {
		cmd.Factory.IOStreams.SetStderrTTY(v)
		return nil
	}
}

func WithEnv(t *testing.T, env map[string]string) ExecuteOptions {
	return func(cmd *root.CmdRoot) error {
		// clean all existing an variables
		originalEnv := os.Environ()
		for _, item := range originalEnv {
			if k, _, ok := strings.Cut(item, "="); ok && strings.HasPrefix(k, "C8Y_") {
				os.Unsetenv(k)
			}
		}

		for k, v := range env {
			t.Setenv(k, v)
		}

		t.Cleanup(func() {
			for _, item := range originalEnv {
				if k, v, ok := strings.Cut(item, "="); ok {
					os.Setenv(k, v)
				}
			}
		})
		return nil
	}
}

// WithSensitiveLogging controls the hiding of sensitive information in the logs and dry run output
func WithSensitiveLogging(t *testing.T, v bool) ExecuteOptions {
	return func(cmd *root.CmdRoot) error {
		key := config.GetEnvKey(config.SettingsLoggerHideSensitive)
		originalValue, found := os.LookupEnv(key)
		os.Setenv(key, fmt.Sprintf("%v", v))
		t.Cleanup(func() {
			if found {
				os.Setenv(key, originalValue)
			} else {
				os.Unsetenv(key)
			}
		})
		return nil
	}
}

func WithOSStdIn(t *testing.T, v string) ExecuteOptions {
	return func(cmd *root.CmdRoot) error {
		stdin := fakestdin.NewStdIn()
		stdin.Write(v)
		t.Cleanup(stdin.Restore)
		return nil
	}
}

func WithStdIn(v string) ExecuteOptions {
	return func(cmd *root.CmdRoot) error {
		stdin := bytes.NewBufferString(v)
		cmd.SetIn(stdin)
		cmd.Factory.IOStreams.In = io.NopCloser(stdin)
		return nil
	}
}

func WithCaptureStdOut(w io.Writer) ExecuteOptions {
	return func(cmd *root.CmdRoot) error {
		cmd.SetOut(w)
		cmd.Factory.IOStreams.Out = w
		consoler, err := cmd.Factory.Console()
		if err != nil {
			panic(err)
		}
		consoler.SetOut(w)
		return nil
	}
}

func WithCaptureStdErr(w io.Writer) ExecuteOptions {
	return func(cmd *root.CmdRoot) error {
		cmd.SetErr(w)
		cmd.Factory.IOStreams.ErrOut = w
		return nil
	}
}
