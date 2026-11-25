// Package smartgroups wires the `c8y smartgroups` command and its subcommands.
// All subcommands are hand-written v2 c8ystream commands (backed by the go-c8y v2
// SmartGroups service), so this group command is itself hand-written rather than
// generated.
package smartgroups

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdSmartgroups struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdSmartgroups {
	ccmd := &SubCmdSmartgroups{}

	cmd := &cobra.Command{
		Use:   "smartgroups",
		Short: "Cumulocity smart groups",
		Long:  `REST endpoint to interact with Cumulocity smart groups. A smart group is an inventory managed object and can also be managed via the Inventory api.`,
	}

	// Subcommands
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewListCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
