package users

import (
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/users/list"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdUsers struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdUsers {
	ccmd := &SubCmdUsers{}

	cmd := &cobra.Command{
		Use:   "users",
		Short: "Cumulocity IoT DataHub Users",
		Long:  `Cumulocity IoT DataHub Users`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
