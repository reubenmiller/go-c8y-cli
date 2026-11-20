// Package measurements wires the `c8y measurements` command and its spec-derived
// subcommands. Every one is a hand-written v2 c8ystream command, so this group
// command is itself hand-written rather than generated. The CLI-only measurements
// subcommands (subscribe/assert/createBulk) are attached on top of this command
// in pkg/cmd/root.
package measurements

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// SubCmdMeasurements is the `measurements` group command.
type SubCmdMeasurements struct {
	*subcommand.SubCommand
}

// NewSubCommand builds the `measurements` group command and attaches its
// spec-derived subcommands.
func NewSubCommand(f *cmdutil.Factory) *SubCmdMeasurements {
	ccmd := &SubCmdMeasurements{}

	cmd := &cobra.Command{
		Use:   "measurements",
		Short: "Cumulocity measurements",
		Long:  `REST endpoint to interact with Cumulocity measurements`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetSeriesCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCollectionCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
