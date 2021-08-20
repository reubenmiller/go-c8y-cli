package assert

import (
	cmdExists "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/assert/exists"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAssert struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdAssert {
	ccmd := &SubCmdAssert{}

	cmd := &cobra.Command{
		Use:   "assert",
		Short: "Cumulocity inventory assertions",
		Long:  `REST endpoint to interact with Cumulocity managed objects`,
	}

	// Subcommands
	cmd.AddCommand(cmdExists.NewCmdExists(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
