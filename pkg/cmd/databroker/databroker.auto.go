package cmd

import (
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/databroker/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/databroker/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/databroker/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdDatabroker struct {
	*subcommand.SubCommand
}

func NewSubCmdDatabroker(f *cmdutil.Factory) *SubCmdDatabroker {
	ccmd := &SubCmdDatabroker{}

	cmd := &cobra.Command{
		Use:   "databroker",
		Short: "Cumulocity databroker",
		Long:  `REST endpoint to interact with Cumulocity databroker`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
