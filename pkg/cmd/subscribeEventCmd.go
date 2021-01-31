// TODO

package cmd

import (
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type subscribeEventCmd struct {
	*baseCmd

	flagDurationSec int64
	flagCount       int64
}

func NewSubscribeEventCmd() *subscribeEventCmd {
	ccmd := &subscribeEventCmd{}

	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to realtime events",
		Long:  `Subscribe to realtime events`,
		Example: `
$ c8y events subscribe --device 12345
Subscribe to events (in realtime) for device 12345

$ c8y events subscribe --device 12345 --duration 30
Subscribe to events (in realtime) for device 12345 for 30 seconds

$ c8y events subscribe --count 10
Subscribe to events (in realtime) for all devices, and stop after receiving 10 events
		`,
		RunE: ccmd.subscribeEvent,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device ID")
	cmd.Flags().Int64Var(&ccmd.flagDurationSec, "duration", 30, "Timeout in seconds")
	cmd.Flags().Int64Var(&ccmd.flagCount, "count", 0, "Max number of realtime notifications to wait for")

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *subscribeEventCmd) subscribeEvent(cmd *cobra.Command, args []string) error {

	// options
	device := "*"

	if cmd.Flags().Changed("device") {
		deviceInputValues, deviceValue, err := getFormattedDeviceSlice(cmd, args, "device")

		if err != nil {
			return newUserError("no matching devices found", deviceInputValues, err)
		}

		if len(deviceValue) == 0 {
			return newUserError("no matching devices found", deviceInputValues)
		}

		for _, item := range deviceValue {
			if item != "" {
				device = newIDValue(item).GetID()
			}
		}
	}

	return subscribe(c8y.RealtimeEvents(device), n.flagDurationSec, n.flagCount, cmd)
}
