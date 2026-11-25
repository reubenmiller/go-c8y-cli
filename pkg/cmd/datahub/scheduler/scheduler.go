// Package scheduler wires the `c8y datahub scheduler` command and its
// subcommands (list). All are hand-written v2 c8ystream commands, so this
// group command is itself hand-written rather than generated.
package scheduler

import (
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
		Use:   "scheduler",
		Short: "Cumulocity DataHub scheduler",
		Long:  `Cumulocity DataHub scheduler`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
