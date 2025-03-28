package applications

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	cmdDisable "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/tenants/applications/disable"
	cmdEnable "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/tenants/applications/enable"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/tenants/applications/list"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdApplications struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdApplications {
	ccmd := &SubCmdApplications{}

	cmd := &cobra.Command{
		Use:   "applications",
		Short: "Manage applications by tenant",
		Long:  `Manage the applications used by sub-tenants`,
	}

	// Subcommands
	cmd.AddCommand(cmdEnable.NewEnableCmd(f).GetCommand())
	cmd.AddCommand(cmdDisable.NewDisableCmd(f).GetCommand())
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
