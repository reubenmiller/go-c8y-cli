package factory

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/desiredstate"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type StateChecker interface {
	GetStateHandler(*cobra.Command, *c8y.Client) desiredstate.StateDefiner
	GetValue(interface{}, interface{}) []byte
}

func NewAssertCmdFactory(cmd *cobra.Command, f *cmdutil.Factory, h StateChecker) *cobra.Command {
	cmd.Flags().StringSlice("id", []string{""}, "Inventory id (required) (accepts pipeline)")
	cmd.Flags().String("duration", "30s", "Timeout duration. i.e. 30s or 1m (1 minute)")
	cmd.Flags().String("interval", "5s", "Interval to check on the status, i.e. 10s or 1min")
	cmd.Flags().Int64("retries", 0, "Number of retries before giving up per id")
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
		consol, err := f.Console()
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

		retries, err := cmd.Flags().GetInt64("retries")
		if err != nil {
			return err
		}

		strictMode, err := cmd.Flags().GetBool("strict")
		if err != nil {
			return err
		}

		// path parameters
		path := flags.NewStringTemplate("{id}")
		err = flags.WithPathParameters(
			cmd,
			path,
			inputIterators,
			c8yfetcher.WithIDSlice(args, "id", "id"),
		)
		if err != nil {
			return err
		}

		state := h.GetStateHandler(cmd, client)

		totalErrors := 0
		var lastErr error
		var result interface{}

		for {
			itemID, input, inputErr := path.Execute(false)

			if inputErr == io.EOF {
				break
			}

			if totalErrors >= cfg.AbortOnErrorCount() {
				msg := fmt.Sprintf("Too many errors. total=%d, max=%d. lastError=%s", totalErrors, cfg.AbortOnErrorCount(), lastErr)
				return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitAbortedWithErrors, msg)
			}

			if inputErr == nil {
				// Skip checking if the input has errors
				_ = state.SetValue(itemID)
				result, err = desiredstate.WaitForWithRetries(retries, interval, duration, state)
				if err == nil {
					outValue := h.GetValue(result, input)

					if jsonUtilities.IsJSONObject(outValue) {
						_ = f.WriteJSONToConsole(cfg, cmd, "", outValue)
					} else {
						fmt.Fprintf(consol, "%s\n", outValue)
					}
				}
			} else {
				err = inputErr
			}

			if err != nil {
				if !errors.Is(err, cmderrors.ErrAssertion) || strictMode {
					totalErrors++
					lastErr = f.CheckPostCommandError(err)

					// wrap error so it is not printed twice, and is still an assertion error
					cErr := cmderrors.NewUserErrorWithExitCode(cmderrors.ExitAssertionError, lastErr)
					cErr.Processed = true
					lastErr = cErr
				}
			}
		}
		if totalErrors > 0 {
			return lastErr
		}
		return nil
	}
	return cmd
}

func NewAssertDeviceCmdFactory(cmd *cobra.Command, f *cmdutil.Factory, h StateChecker) *cobra.Command {
	cmd.Flags().StringSlice("device", []string{""}, "The ManagedObject which is the source of this event. (accepts pipeline)")
	cmd.Flags().String("duration", "30s", "Timeout duration. i.e. 30s or 1m (1 minute)")
	cmd.Flags().String("interval", "5s", "Interval to check on the status, i.e. 10s or 1min")
	cmd.Flags().Int64("retries", 0, "Number of retries before giving up per id")
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
		consol, err := f.Console()
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

		retries, err := cmd.Flags().GetInt64("retries")
		if err != nil {
			return err
		}

		strictMode, err := cmd.Flags().GetBool("strict")
		if err != nil {
			return err
		}

		// path parameters
		path := flags.NewStringTemplate("{device}")
		err = flags.WithPathParameters(
			cmd,
			path,
			inputIterators,
			c8yfetcher.WithDeviceByNameFirstMatch(client, args, "device", "device"),
		)
		if err != nil {
			return err
		}

		state := h.GetStateHandler(cmd, client)

		totalErrors := 0
		var lastErr error
		var result interface{}

		for {
			itemID, input, inputErr := path.Execute(false)

			if inputErr == io.EOF {
				break
			}

			if totalErrors >= cfg.AbortOnErrorCount() {
				msg := fmt.Sprintf("Too many errors. total=%d, max=%d. lastError=%s", totalErrors, cfg.AbortOnErrorCount(), lastErr)
				return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitAbortedWithErrors, msg)
			}

			if inputErr == nil {
				// Skip checking if the input has errors
				_ = state.SetValue(itemID)
				result, err = desiredstate.WaitForWithRetries(retries, interval, duration, state)
				if err == nil {
					outValue := h.GetValue(result, input)

					if jsonUtilities.IsJSONObject(outValue) {
						_ = f.WriteJSONToConsole(cfg, cmd, "", outValue)
					} else {
						fmt.Fprintf(consol, "%s\n", outValue)
					}
				}
			} else {
				err = inputErr
			}

			if err != nil {
				if !errors.Is(err, cmderrors.ErrAssertion) || strictMode {
					totalErrors++
					lastErr = f.CheckPostCommandError(err)

					// wrap error so it is not printed twice, and is still an assertion error
					cErr := cmderrors.NewUserErrorWithExitCode(cmderrors.ExitAssertionError, lastErr)
					cErr.Processed = true
					lastErr = cErr
				}
			}
		}

		// TODO: In strict mode print out the error earlier
		if totalErrors > 0 {
			return lastErr
		}
		return nil
	}
	return cmd
}
