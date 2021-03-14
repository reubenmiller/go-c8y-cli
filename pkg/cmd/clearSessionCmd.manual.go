package cmd

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/spf13/cobra"
)

// ClearSessionCmd clear session command
type ClearSessionCmd struct {
	Shell string

	*baseCmd
}

// NewClearSessionCmd creates a command used to clear the current session
func NewClearSessionCmd() *ClearSessionCmd {
	ccmd := &ClearSessionCmd{}

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear current session",
		Long:  `Clear the current session by returning all the environment variables which should be unset via a shell snippet`,
		Example: heredoc.Doc(`
### Example 1: Clear session in bash
$ eval $(c8y session clear)

Clear the current session

## Example 2: Clear session in fish
$ c8y session clear | source
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true
	cmd.Flags().StringVar(&ccmd.Shell, "shell", "bash", "Shell type")
	_ = cmd.MarkFlagRequired("shell")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("shell", "bash", "zsh", "fish", "powershell"),
	)
	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *ClearSessionCmd) RunE(cmd *cobra.Command, args []string) error {
	shell := ShellBash
	clearEnvironmentVariables(shell.FromString(n.Shell))
	return nil
}
