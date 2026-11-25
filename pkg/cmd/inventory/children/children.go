// Package children wires the `c8y inventory children` command and its
// subcommands. All subcommands are hand-written v2 c8ystream commands that
// dispatch the --childType flag to the go-c8y v2 ManagedObjects child-reference
// services, so this group command is itself hand-written rather than generated.
package children

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdChildren struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdChildren {
	ccmd := &SubCmdChildren{}

	cmd := &cobra.Command{
		Use:   "children",
		Short: "Cumulocity managed object child references",
		Long:  `Manage child entities (assets, additions and device) for managed objects`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewAssignCmd(f).GetCommand())
	cmd.AddCommand(NewUnassignCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
