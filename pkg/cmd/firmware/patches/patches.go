// Package patches wires the `c8y firmware patches` command and its subcommands.
// The subcommands are hand-written v2 c8ystream commands backed by the go-c8y v2
// Repository.Firmware.Patches service. The create command is attached in
// pkg/cmd/root/root.go.
package patches

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdPatches struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdPatches {
	ccmd := &SubCmdPatches{}

	cmd := &cobra.Command{
		Use:   "patches",
		Short: "Cumulocity firmware patch management",
		Long:  `Firmware patch management to create/list/delete patches`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
