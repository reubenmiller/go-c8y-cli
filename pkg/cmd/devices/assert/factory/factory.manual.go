package factory

import (
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/desiredstate"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/worker"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type StateChecker interface {
	GetStateHandler(*cobra.Command, *c8y.Client) desiredstate.StateDefiner
	GetValue(interface{}, interface{}) []byte
}

func NewAssertCmdFactory(cmd *cobra.Command, f *cmdutil.Factory, h StateChecker) *cobra.Command {
	cmd.Flags().StringSlice("id", []string{""}, "Inventory id (required) (accepts pipeline)")
	cmd.Flags().String("duration", "30s", "Timeout duration. i.e. 30s or 1m (1 minute)")
	cmd.Flags().String("interval", "5s", "Interval to check on the status, i.e. 10s or 1min")
	cmd.Flags().Int64("attempts", -1, "Number of attempts before giving up per id (-1 = unlimited)")
	cmd.Flags().Bool("strict", false, "Strict mode, fail if no match is found")
	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true, "deviceId", "source.id", "managedObject.id", "id"),
	)

	cmd.SilenceUsage = true

	completion.WithOptions(
		cmd,
	)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := f.Config()
		if err != nil {
			return err
		}
		client, err := f.Client()
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

		interval, err := flags.GetDurationFlag(cmd, "interval", true, time.Second)
		if err != nil {
			return err
		}

		attempts, err := cmd.Flags().GetInt64("attempts")
		if err != nil {
			return err
		}

		strictMode, err := cmd.Flags().GetBool("strict")
		if err != nil {
			return err
		}

		// path parameters
		path := flags.NewStringTemplate("{id}")
		err = flags.WithParameters(
			cmd,
			path,
			inputIterators,
			c8yfetcher.WithIDSlice(args, "id", "id"),
		)
		if err != nil {
			return err
		}

		return f.RunWithGenericWorkers(cmd, inputIterators, path, func(j worker.Job) (any, error) {
			state := h.GetStateHandler(cmd, client)

			itemID := gjson.ParseBytes(j.Value.([]byte))
			input := j.Input

			// Skip checking if the input has errors
			_ = state.SetValue(itemID.String())
			result, err := desiredstate.WaitForWithRetries(attempts, interval, duration, state)
			if err == nil {
				outValue := h.GetValue(result, input)
				_ = f.WriteOutputWithoutPropertyGuess(outValue, cmdutil.OutputContext{})
			}
			return nil, err
		}, cmdutil.ProcessAssertError(f, strictMode))
	}
	return cmd
}

func NewAssertDeviceCmdFactory(cmd *cobra.Command, f *cmdutil.Factory, h StateChecker) *cobra.Command {
	cmd.Flags().StringSlice("device", []string{""}, "The ManagedObject which is the source of this event. (accepts pipeline)")
	cmd.Flags().String("duration", "30s", "Timeout duration. i.e. 30s or 1m (1 minute)")
	cmd.Flags().String("interval", "5s", "Interval to check on the status, i.e. 10s or 1min")
	cmd.Flags().Int64("attempts", -1, "Number of attempts before giving up per id (-1 = unlimited)")
	cmd.Flags().Bool("strict", false, "Strict mode, fail if no match is found")
	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "device", true, "deviceId", "source.id", "managedObject.id", "id"),
	)

	cmd.SilenceUsage = true

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return f.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "device", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := f.Config()
		if err != nil {
			return err
		}
		client, err := f.Client()
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

		interval, err := flags.GetDurationFlag(cmd, "interval", true, time.Second)
		if err != nil {
			return err
		}

		attempts, err := cmd.Flags().GetInt64("attempts")
		if err != nil {
			return err
		}

		strictMode, err := cmd.Flags().GetBool("strict")
		if err != nil {
			return err
		}

		// path parameters
		path := flags.NewStringTemplate("{device}")
		err = flags.WithParameters(
			cmd,
			path,
			inputIterators,
			c8yfetcher.WithDeviceByNameFirstMatch(f, args, "device", "device"),
		)
		if err != nil {
			return err
		}

		return f.RunWithGenericWorkers(cmd, inputIterators, path, func(j worker.Job) (any, error) {
			state := h.GetStateHandler(cmd, client)

			itemID := gjson.ParseBytes(j.Value.([]byte))
			input := j.Input

			// Skip checking if the input has errors
			_ = state.SetValue(itemID.String())
			result, err := desiredstate.WaitForWithRetries(attempts, interval, duration, state)
			if err == nil {
				outValue := h.GetValue(result, input)
				_ = f.WriteOutputWithoutPropertyGuess(outValue, cmdutil.OutputContext{})
			}
			return nil, err
		}, cmdutil.ProcessAssertError(f, strictMode))
	}
	return cmd
}
