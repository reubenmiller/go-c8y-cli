package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenants/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenants/delete"
	cmdDisableApplication "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenants/disableapplication"
	cmdEnableApplication "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenants/enableapplication"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenants/get"
	cmdGetCurrentTenant "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenants/getcurrenttenant"
	cmdGetVersion "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenants/getversion"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenants/list"
	cmdListReferences "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenants/listreferences"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenants/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdTenants struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdTenants {
	ccmd := &SubCmdTenants{}

	cmd := &cobra.Command{
		Use:   "tenants",
		Short: "Cumulocity tenant",
		Long:  `REST endpoint to interact with Cumulocity tenants`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdGetCurrentTenant.NewGetCurrentTenantCmd(f).GetCommand())
	cmd.AddCommand(cmdEnableApplication.NewEnableApplicationCmd(f).GetCommand())
	cmd.AddCommand(cmdDisableApplication.NewDisableApplicationCmd(f).GetCommand())
	cmd.AddCommand(cmdListReferences.NewListReferencesCmd(f).GetCommand())
	cmd.AddCommand(cmdGetVersion.NewGetVersionCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
