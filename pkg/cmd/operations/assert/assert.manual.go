package assert

import (
	cmdCount "github.com/reubenmiller/go-c8y-cli/pkg/cmd/operations/assert/count"
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
		Short: "Operation assertions",
		Long:  `Assertions related to Cumulocity operations`,
	}

	// Subcommands
	cmd.AddCommand(cmdCount.NewCmdCount(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
