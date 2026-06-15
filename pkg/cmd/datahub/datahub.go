// Package datahub wires the `c8y datahub` command and its subcommands (query +
// the jobs subgroup). All are hand-written v2 c8ystream commands, so this group
// command is itself hand-written rather than generated.
package datahub

import (
	cmdJobs "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/jobs"
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
		Short: "Cumulocity Data Hub api",
		Long:  `Data Hub api`,
	}

	// Subcommands
	cmd.AddCommand(cmdQuery.NewQueryCmd(f).GetCommand())
	cmd.AddCommand(cmdJobs.NewSubCommand(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
