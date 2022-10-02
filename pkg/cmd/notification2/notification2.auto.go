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
