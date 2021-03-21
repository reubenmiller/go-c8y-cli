package subscribeall

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ysubscribe"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdSubscribeAll struct {
	flagDurationSec int64
	flagCount       int64

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdSubscribeAll(f *cmdutil.Factory) *CmdSubscribeAll {
	ccmd := &CmdSubscribeAll{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "subscribeAll",
		Short: "Subscribe to all realtime notifications",
		Long:  `Subscribe to all realtime notifications`,
		Example: heredoc.Doc(`
$ c8y realtime subscribeAll --device 12345 --duration 90

Subscribe to all notifications (alarms/events/operations etc.) for device 12345 for 90 seconds
		`),
		RunE: ccmd.RunE,
	}

	// Flags
	cmd.Flags().String("device", "", "Device ID")
	cmd.Flags().Int64Var(&ccmd.flagDurationSec, "duration", 30, "Timeout in seconds")
	cmd.Flags().Int64Var(&ccmd.flagCount, "count", 0, "Max number of realtime notifications to wait for")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdSubscribeAll) RunE(cmd *cobra.Command, args []string) error {
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	// path parameters
	path := flags.NewStringTemplate("{device}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		flags.WithStringDefaultValue("*", "device", "device"),
		c8yfetcher.WithDeviceByNameFirstMatch(client, args, "device", "device"),
	)
	if err != nil {
		return err
	}

	device, _, err := path.Execute(false)
	if err != nil {
		return err
	}

	patterns := []string{
		c8y.RealtimeAlarms(device),
		c8y.RealtimeEvents(device),
		c8y.RealtimeMeasurements(device),
		c8y.RealtimeOperations(device),
	}

	return c8ysubscribe.SubscribeMultiple(client, log, patterns, n.flagDurationSec, n.flagCount, false, cmd)
}
