// Package jobs wires the `c8y datahub jobs` command and its subcommands
// (create/get/cancel/listResults). All are hand-written v2 c8ystream commands,
// so this group command is itself hand-written rather than generated.
package jobs

import (
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
		Short: "Cumulocity DataHub Jobs",
		Long:  `Cumulocity DataHub Jobs`,
	}

	// Subcommands
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewCancelCmd(f).GetCommand())
	cmd.AddCommand(NewListResultsCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
