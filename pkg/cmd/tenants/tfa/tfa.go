// Package tfa wires the `c8y tenants tfa` command and its subcommands
// (get/update). Both are hand-written v2 c8ystream commands, so this group
// command is itself hand-written rather than generated.
package tfa

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/tenants/tfa/get"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/tenants/tfa/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdTfa struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdTfa {
	ccmd := &SubCmdTfa{}

	cmd := &cobra.Command{
		Use:   "tfa",
		Short: "Cumulocity Tenant Two-Factor-Authentication Setting",
		Long:  `Managed the Two-Factor-Authentication settings used by a tenant`,
	}

	// Subcommands
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
