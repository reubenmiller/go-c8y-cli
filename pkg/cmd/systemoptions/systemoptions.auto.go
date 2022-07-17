package systemoptions

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/systemoptions/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/systemoptions/list"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdSystemoptions struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdSystemoptions {
	ccmd := &SubCmdSystemoptions{}

	cmd := &cobra.Command{
		Use:   "systemoptions",
		Short: "Cumulocity systemOptions",
		Long:  `REST endpoint to interact with Cumulocity systemOptions`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
