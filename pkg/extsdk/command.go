package extsdk

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// NewRootCommand returns a cobra root command for an extension binary. The completion
// command is hidden, and cobra's hidden __complete command is available, so the binary
// works with "binary.completion: cobra" in extension.yaml
func NewRootCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           name,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.CompletionOptions.HiddenDefaultCmd = true
	return cmd
}

// Execute runs the root command, prints any error to stderr, and exits with a non-zero
// exit code on errors. The exit code of a failed c8y command (see Exec) is preserved
func Execute(root *cobra.Command) {
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() > 0 {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}
