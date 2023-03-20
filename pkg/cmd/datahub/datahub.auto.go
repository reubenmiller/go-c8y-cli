package datahub

import (
	cmdQuery "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/query"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdDatahub struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdDatahub {
	ccmd := &SubCmdDatahub{}

	cmd := &cobra.Command{
		Use:   "datahub",
		Short: "Cumulocity IoT Data Hub api",
		Long:  `Data Hub api`,
	}

	// Subcommands
	cmd.AddCommand(cmdQuery.NewQueryCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
