// Package subscriptions wires the `c8y notification2 subscriptions` command and
// its subcommands (list/create/get/deleteBySource/delete). All are hand-written
// v2 c8ystream commands backed by the go-c8y v2 Notification2 service. The
// CLI-only `subscribe` command is attached on top of this command in
// pkg/cmd/root.
package subscriptions

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdSubscriptions struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdSubscriptions {
	ccmd := &SubCmdSubscriptions{}

	cmd := &cobra.Command{
		Use:   "subscriptions",
		Short: "Cumulocity notification2 subscriptions",
		Long:  `Methods to create, retrieve and delete notification subscriptions`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteBySourceCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
