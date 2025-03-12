package tenants

import (
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/features/tenants/delete"
	cmdDisable "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/features/tenants/disable"
	cmdEnable "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/features/tenants/enable"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/features/tenants/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/features/tenants/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdTenants struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdTenants {
	ccmd := &SubCmdTenants{}

	cmd := &cobra.Command{
		Use:   "tenants",
		Short: "Cumulocity tenant features from the management tenant",
		Long:  `Managed feature toggles from the management tenant`,
	}

	// Subcommands
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdEnable.NewEnableCmd(f).GetCommand())
	cmd.AddCommand(cmdDisable.NewDisableCmd(f).GetCommand())
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
