package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type realtimeCmd struct {
	*subcommand.SubCommand
}

func NewRealtimeCmd() *realtimeCmd {
	ccmd := &realtimeCmd{}

	cmd := &cobra.Command{
		Use:   "realtime",
		Short: "Cumulocity realtime notifications",
		Long:  `Cumulocity realtime notifications`,
	}

	// Subcommands
	cmd.AddCommand(newSubscribeRealtimeCmd().GetCommand())
	cmd.AddCommand(newSubscribeAllRealtimeCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
