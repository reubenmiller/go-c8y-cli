// Package availability wires the `c8y devices availability` command and its
// subcommands. The get/set subcommands are hand-written v2 c8ystream commands
// backed by the go-c8y v2 ManagedObjects availability endpoint and a managed-
// object update, so this group command is itself hand-written rather than
// generated.
package availability

import (
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/availability/get"
	cmdSet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/availability/set"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAvailability struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdAvailability {
	ccmd := &SubCmdAvailability{}

	cmd := &cobra.Command{
		Use:   "availability",
		Short: "Cumulocity device availability",
		Long:  `REST endpoint to interact with Cumulocity devices`,
	}

	// Subcommands
	cmd.AddCommand(cmdSet.NewSetCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
