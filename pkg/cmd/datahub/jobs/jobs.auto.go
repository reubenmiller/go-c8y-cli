package jobs

import (
	cmdCancel "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/jobs/cancel"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/jobs/create"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/jobs/get"
	cmdListResults "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/jobs/listresults"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdJobs struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdJobs {
	ccmd := &SubCmdJobs{}

	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Cumulocity IoT DataHub Jobs",
		Long:  `Cumulocity IoT DataHub Jobs`,
	}

	// Subcommands
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdCancel.NewCancelCmd(f).GetCommand())
	cmd.AddCommand(cmdListResults.NewListResultsCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
