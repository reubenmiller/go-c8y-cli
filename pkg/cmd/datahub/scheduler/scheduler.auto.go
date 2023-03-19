package scheduler

import (
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/scheduler/list"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdScheduler struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdScheduler {
	ccmd := &SubCmdScheduler{}

	cmd := &cobra.Command{
		Use:    "scheduler",
		Short:  "Cumulocity IoT DataHub Scheduler",
		Long:   `Cumulocity IoT DataHub Scheduler`,
		Hidden: true,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
