// Package inventory wires the `c8y inventory` command and its CRUD/query
// subcommands. All subcommands are hand-written v2 c8ystream commands backed by
// the go-c8y v2 ManagedObjects service, so this group command is itself
// hand-written rather than generated.
package inventory

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdInventory struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdInventory {
	ccmd := &SubCmdInventory{}

	cmd := &cobra.Command{
		Use:   "inventory",
		Short: "Cumulocity managed objects",
		Long:  `REST endpoint to interact with Cumulocity managed objects`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewCountCmd(f).GetCommand())
	cmd.AddCommand(NewFindByTextCmd(f).GetCommand())
	cmd.AddCommand(NewFindCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
