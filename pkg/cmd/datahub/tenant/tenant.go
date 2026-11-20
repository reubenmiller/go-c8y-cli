// Package tenant wires the `c8y datahub tenant` command and its subcommands
// (get). All are hand-written v2 c8ystream commands, so this group command is
// itself hand-written rather than generated.
package tenant

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdTenant struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdTenant {
	ccmd := &SubCmdTenant{}

	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Cumulocity DataHub tenant",
		Long:  `Cumulocity DataHub tenant`,
	}

	// Subcommands
	cmd.AddCommand(NewGetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
