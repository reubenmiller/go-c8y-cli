package cmd

import (
	"github.com/spf13/cobra"
)

type realtimeCmd struct {
	*baseCmd
}

func NewRealtimeCmd() *realtimeCmd {
	ccmd := &realtimeCmd{}

	cmd := &cobra.Command{
		Use:   "realtime",
		Short: "Realtime notifications",
		Long:  `Realtime notifications`,
	}

	// Subcommands
	cmd.AddCommand(newSubscribeRealtimeCmd().getCommand())
	cmd.AddCommand(newSubscribeAllRealtimeCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
