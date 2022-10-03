package parents

import (
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/parents/get"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdParents struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdParents {
	ccmd := &SubCmdParents{}

	cmd := &cobra.Command{
		Use:   "parents",
		Short: "Get parent managed objects",
		Long:  `Get parent managed objects such as addition, asset, or device parents`,
	}

	// Subcommands
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
