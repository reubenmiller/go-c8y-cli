package cmd

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

type subscribeRealtimeCmd struct {
	flagChannel     string
	flagDurationSec int64
	flagCount       int64

	*baseCmd
}

func newSubscribeRealtimeCmd() *subscribeRealtimeCmd {
	ccmd := &subscribeRealtimeCmd{}

	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to realtime notifications",
		Long:  `Subscribe to realtime notifications`,
		Example: heredoc.Doc(`
$ c8y realtime subscribe --channel /measurements/* --duration 90

Subscribe to all measurements for 90 seconds
		`),
		RunE: ccmd.subscribeRealtime,
	}

	// Flags
	cmd.Flags().StringVar(&ccmd.flagChannel, "channel", "", "Channel name i.e. /measurements/12345 or /measurements/*")
	cmd.Flags().Int64Var(&ccmd.flagDurationSec, "duration", 30, "Timeout in seconds")
	cmd.Flags().Int64Var(&ccmd.flagCount, "count", 0, "Max number of realtime notifications to wait for")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *subscribeRealtimeCmd) subscribeRealtime(cmd *cobra.Command, args []string) error {
	return subscribe(n.flagChannel, n.flagDurationSec, n.flagCount, cmd)
}
