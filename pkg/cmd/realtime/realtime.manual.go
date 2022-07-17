package realtime

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/realtime/subscribe"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/realtime/subscribeall"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdRealtime struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdRealtime {
	ccmd := &SubCmdRealtime{}

	cmd := &cobra.Command{
		Use:   "realtime",
		Short: "Cumulocity realtime notifications",
		Long:  `Cumulocity realtime notifications`,
	}

	// Subcommands
	cmd.AddCommand(subscribe.NewCmdSubscribe(f).GetCommand())
	cmd.AddCommand(subscribeall.NewCmdSubscribeAll(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
