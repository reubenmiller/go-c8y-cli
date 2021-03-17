package cmd

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type subscribeRealtimeAllCmd struct {
	flagDurationSec int64
	flagCount       int64

	*subcommand.SubCommand
}

func newSubscribeAllRealtimeCmd() *subscribeRealtimeAllCmd {
	ccmd := &subscribeRealtimeAllCmd{}

	cmd := &cobra.Command{
		Use:   "subscribeAll",
		Short: "Subscribe to all realtime notifications",
		Long:  `Subscribe to all realtime notifications`,
		Example: heredoc.Doc(`
$ c8y realtime subscribeAll --device 12345 --duration 90

Subscribe to all notifications (alarms/events/operations etc.) for device 12345 for 90 seconds
		`),
		RunE: ccmd.subscribeAllRealtime,
	}

	// Flags
	cmd.Flags().StringSlice("device", []string{""}, "Device ID")
	cmd.Flags().Int64Var(&ccmd.flagDurationSec, "duration", 30, "Timeout in seconds")
	cmd.Flags().Int64Var(&ccmd.flagCount, "count", 0, "Max number of realtime notifications to wait for")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *subscribeRealtimeAllCmd) subscribeAllRealtime(cmd *cobra.Command, args []string) error {
	device := "*"

	if cmd.Flags().Changed("device") {
		deviceInputValues, deviceValue, err := getFormattedDeviceSlice(cmd, args, "device")

		if err != nil {
			return cmderrors.NewUserError("no matching devices found", deviceInputValues, err)
		}

		if len(deviceValue) == 0 {
			return cmderrors.NewUserError("no matching devices found", deviceInputValues)
		}

		for _, item := range deviceValue {
			if item != "" {
				device = newIDValue(item).GetID()
			}
		}
	}

	patterns := []string{
		c8y.RealtimeAlarms(device),
		c8y.RealtimeEvents(device),
		c8y.RealtimeMeasurements(device),
		c8y.RealtimeOperations(device),
	}

	return subscribeMultiple(patterns, n.flagDurationSec, n.flagCount, false, cmd)
}
