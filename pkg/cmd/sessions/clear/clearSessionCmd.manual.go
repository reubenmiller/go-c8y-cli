package clear

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/utilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// CmdClearSession clear session command
type CmdClearSession struct {
	Shell string

	*subcommand.SubCommand

	factory *cmdutil.Factory
	Config  func() (*config.Config, error)
	Client  func() (*c8y.Client, error)
}

// NewCmdClearSession creates a command used to clear the current session
func NewCmdClearSession(f *cmdutil.Factory) *CmdClearSession {
	ccmd := &CmdClearSession{
		factory: f,
		Config:  f.Config,
		Client:  f.Client,
	}

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
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdClearSession) RunE(cmd *cobra.Command, args []string) error {
	shell := utilities.ShellBash
	utilities.ClearEnvironmentVariables(shell.FromString(n.Shell))
	return nil
}
