// Package datahub wires the `c8y datahub` command and its subcommands (query +
// the configuration/jobs/scheduler/tenant/users subgroups). All are
// hand-written v2 c8ystream commands, so this group command is itself
// hand-written rather than generated.
package datahub

import (
	cmdConfiguration "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/configuration"
	cmdJobs "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/jobs"
	cmdScheduler "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/scheduler"
	cmdTenant "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/tenant"
	cmdUsers "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/users"
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
	cmd.AddCommand(NewQueryCmd(f).GetCommand())
	cmd.AddCommand(cmdConfiguration.NewSubCommand(f).GetCommand())
	cmd.AddCommand(cmdJobs.NewSubCommand(f).GetCommand())
	cmd.AddCommand(cmdScheduler.NewSubCommand(f).GetCommand())
	cmd.AddCommand(cmdTenant.NewSubCommand(f).GetCommand())
	cmd.AddCommand(cmdUsers.NewSubCommand(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
