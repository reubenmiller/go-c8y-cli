package assert

import (
	cmdExists "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/assert/exists"
	cmdFragments "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/assert/fragments"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAssert struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdAssert {
	ccmd := &SubCmdAssert{}

	cmd := &cobra.Command{
		Use:   "assert",
		Short: "Cumulocity device assertions",
		Long:  `Device assertions`,
	}

	// Subcommands
	cmd.AddCommand(cmdExists.NewCmdExists(f).GetCommand())
	cmd.AddCommand(cmdFragments.NewCmdFragments(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
