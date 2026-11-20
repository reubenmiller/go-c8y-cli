// Package tenants wires the `c8y features tenants` command and its subcommands
// (list/enable/disable/update/delete). All are hand-written v2 c8ystream commands
// (backed by the go-c8y v2 Features.Tenants service), so this group command is
// itself hand-written rather than generated.
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
		Short: "Cumulocity tenant features from the management tenant",
		Long:  `Managed feature toggles from the management tenant`,
	}

	// Subcommands
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewEnableCmd(f).GetCommand())
	cmd.AddCommand(NewDisableCmd(f).GetCommand())
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
