package clear

import (
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/shell"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/utilities"
	"github.com/spf13/cobra"
)

// CmdClearSession clear session command
type CmdClearSession struct {
	Shell string

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCmdClearSession creates a command used to clear the current session
func NewCmdClearSession(f *cmdutil.Factory) *CmdClearSession {
	ccmd := &CmdClearSession{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear current session",
		Long:  `Clear the current session by returning all the environment variables which should be unset via a shell snippet`,
		Example: heredoc.Doc(`
### Example 1: Clear session in bash
$ eval $(c8y sessions clear)

Clear the current session

## Example 2: Clear session in fish
$ c8y sessions clear | source
		`),
		RunE: ccmd.RunE,
	}

	cmdutil.DisableEncryptionCheck(cmd)

	cmd.SilenceUsage = true
	cmd.Flags().StringVar(&ccmd.Shell, "shell", "auto", "Shell type. Defaults to auto if not printing to terminal")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("shell", "auto", "bash", "zsh", "fish", "powershell"),
	)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdClearSession) RunE(cmd *cobra.Command, args []string) error {
	shellType := utilities.ShellBash
	if strings.EqualFold(n.Shell, "auto") {
		n.Shell = shell.DetectShell("bash")
	}
	c8ysession.ClearEnvironmentVariables(shellType.FromString(n.Shell))
	return nil
}
