package availability

import (
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/availability/get"
	cmdSet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/availability/set"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
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
