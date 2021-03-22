package alias

import (
	"github.com/MakeNowJust/heredoc"
	deleteCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alias/delete"
	listCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alias/list"
	setCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alias/set"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"

	"github.com/spf13/cobra"
)

func NewCmdAlias(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias <command>",
		Short: "Create command shortcuts",
		Long: heredoc.Doc(`
			Aliases can be used to make shortcuts for c8y commands or to compose multiple commands.
			Run "c8y help alias set" to learn more.
		`),
	}

	cmdutil.DisableAuthCheck(cmd)

	cmd.AddCommand(listCmd.NewCmdList(f, nil))
	cmd.AddCommand(setCmd.NewCmdSet(f, nil))
	cmd.AddCommand(deleteCmd.NewCmdDelete(f, nil))

	return cmd
}
