// Package notification2 wires the `c8y notification2` command. Its subgroups
// (subscriptions/tokens) and the CLI-only `subscriptions subscribe` command are
// attached on top of this command in pkg/cmd/root. Every spec-derived
// subcommand is a hand-written v2 c8ystream command (backed by the go-c8y v2
// Notification2 service), so this group command is itself hand-written rather
// than generated.
package notification2

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdNotification2 struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdNotification2 {
	ccmd := &SubCmdNotification2{}

	cmd := &cobra.Command{
		Use:   "notification2",
		Short: "Cumulocity Notification2",
		Long:  `Managed tokens and subscriptions for notifications`,
	}

	// Subcommands

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
