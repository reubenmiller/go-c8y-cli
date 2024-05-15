// TODO

package subscribe

import (
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysubscribe"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdSubscribe struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	flagCount   int64
	actionTypes []string
}

func NewCmdSubscribe(f *cmdutil.Factory) *CmdSubscribe {
	ccmd := &CmdSubscribe{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to realtime alarms",
		Long:  `Subscribe to realtime alarms`,
		Example: heredoc.Doc(`
$ c8y alarms subscribe --device 12345
Subscribe to alarms (in realtime) for device 12345

$ c8y alarms subscribe --device 12345 --duration 30s
Subscribe to alarms (in realtime) for device 12345 for 30 seconds

$ c8y alarms subscribe --count 10
Subscribe to alarms (in realtime) for all devices, and stop after receiving 10 alarms
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("device", "", "Device ID")
	cmd.Flags().String("duration", "30s", "Duration to subscribe for. i.e. 30s, 1m")
	cmd.Flags().Int64Var(&ccmd.flagCount, "count", 0, "Max number of realtime notifications to wait for")
	cmd.Flags().StringSliceVar(&ccmd.actionTypes, "actionTypes", nil, "Filter by realtime action types, i.e. CREATE,UPDATE,DELETE")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("actionTypes", "CREATE", "UPDATE", "DELETE"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdSubscribe) RunE(cmd *cobra.Command, args []string) error {
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	duration, err := flags.GetDurationFlag(cmd, "duration", true, time.Second)
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
		c8yfetcher.WithDeviceByNameFirstMatch(n.factory, args, "device", "device"),
	)
	if err != nil {
		return err
	}

	device, _, err := path.Execute(false)
	if err != nil {
		return err
	}

	opts := c8ysubscribe.Options{
		Timeout:     duration,
		MaxMessages: n.flagCount,
		ActionTypes: n.actionTypes,
		OnMessage: func(msg string) error {
			return n.factory.WriteOutputWithoutPropertyGuess([]byte(msg), cmdutil.OutputContext{})
		},
	}
	return c8ysubscribe.Subscribe(client, log, c8y.RealtimeAlarms(device), opts)
}
