// TODO

package subscribe

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

type CmdSubscribe struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	flagDurationSec int64
	flagCount       int64
}

func NewCmdSubscribe(f *cmdutil.Factory) *CmdSubscribe {
	ccmd := &CmdSubscribe{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to realtime events",
		Long:  `Subscribe to realtime events`,
		Example: heredoc.Doc(`
$ c8y events subscribe --device 12345
Subscribe to events (in realtime) for device 12345

$ c8y events subscribe --device 12345 --duration 30
Subscribe to events (in realtime) for device 12345 for 30 seconds

$ c8y events subscribe --count 10
Subscribe to events (in realtime) for all devices, and stop after receiving 10 events
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("device", "", "Device ID")
	cmd.Flags().Int64Var(&ccmd.flagDurationSec, "duration", 30, "Timeout in seconds")
	cmd.Flags().Int64Var(&ccmd.flagCount, "count", 0, "Max number of realtime notifications to wait for")

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdSubscribe) RunE(cmd *cobra.Command, args []string) error {
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

	return c8ysubscribe.Subscribe(client, log, c8y.RealtimeEvents(device), n.flagDurationSec, n.flagCount, cmd)
}
