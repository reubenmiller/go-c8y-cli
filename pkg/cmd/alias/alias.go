package alias

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"

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

	cmd.AddCommand(NewCmdList(f, nil))
	cmd.AddCommand(NewCmdSet(f, nil))
	cmd.AddCommand(NewCmdDelete(f, nil))

	return cmd
}
