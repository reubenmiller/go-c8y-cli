// TODO

package cmd

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type subscribeMeasurementCmd struct {
	*baseCmd

	flagDurationSec int64
	flagCount       int64
}

func NewSubscribeMeasurementCmd() *subscribeMeasurementCmd {
	ccmd := &subscribeMeasurementCmd{}

	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to realtime measurements",
		Long:  `Subscribe to realtime measurements`,
		Example: heredoc.Doc(`
$ c8y measurements subscribe --device 12345
Subscribe to measurements (in realtime) for device 12345

$ c8y measurements subscribe --device 12345 --duration 30
Subscribe to measurements (in realtime) for device 12345 for 30 seconds

$ c8y measurements subscribe --count 10
Subscribe to measurements (in realtime) for all devices, and stop after receiving 10 measurements
		`),
		RunE: ccmd.subscribeMeasurement,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device ID")
	cmd.Flags().Int64Var(&ccmd.flagDurationSec, "duration", 30, "Timeout in seconds")
	cmd.Flags().Int64Var(&ccmd.flagCount, "count", 0, "Max number of realtime notifications to wait for")

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *subscribeMeasurementCmd) subscribeMeasurement(cmd *cobra.Command, args []string) error {

	// options
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

	return subscribe(c8y.RealtimeMeasurements(device), n.flagDurationSec, n.flagCount, cmd)
}
