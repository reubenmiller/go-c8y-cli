// Package tenants wires the `c8y tenants` command group. All subcommands are
// hand-written v2 c8ystream commands backed by the go-c8y v2 Tenants service.
package tenants

import (
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
		Short: "Cumulocity tenant",
		Long:  `REST endpoint to interact with Cumulocity tenants`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewEnableApplicationCmd(f).GetCommand())
	cmd.AddCommand(NewDisableApplicationCmd(f).GetCommand())
	cmd.AddCommand(NewListReferencesCmd(f).GetCommand())
	cmd.AddCommand(NewEnableCmd(f).GetCommand())
	cmd.AddCommand(NewDisableCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
