package configuration

import (
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/configuration/list"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdConfiguration struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdConfiguration {
	ccmd := &SubCmdConfiguration{}

	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Cumulocity IoT DataHub Configurations",
		Long:  `Cumulocity IoT DataHub Configurations`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
