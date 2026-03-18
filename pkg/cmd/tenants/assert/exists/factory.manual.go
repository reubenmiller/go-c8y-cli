package exists

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
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type StateChecker interface {
	GetStateHandler(*cobra.Command, *c8y.Client) desiredstate.StateDefiner
	GetValue(interface{}, interface{}) []byte
}

func NewAssertTenantCmdFactory(cmd *cobra.Command, f *cmdutil.Factory, h StateChecker) *cobra.Command {
	cmd.Flags().StringSlice("tenant", []string{}, "Tenant id (accepts pipeline)")
	cmd.Flags().String("duration", "30s", "Timeout duration. i.e. 30s or 1m (1 minute)")
	cmd.Flags().String("interval", "5s", "Interval to check on the status, i.e. 10s or 1min")
	cmd.Flags().Int64("attempts", -1, "Number of attempts before giving up per id (-1 = unlimited)")
	cmd.Flags().Bool("strict", false, "Strict mode, fail if no match is found")
	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("tenant", "tenant", false, "tenantId", "id", "owner.tenant.id", "tenant"),
	)

	cmd.SilenceUsage = true

	completion.WithOptions(
		cmd,
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return f.Client() }),
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
		path := flags.NewStringTemplate("{tenant}")
		if v := f.GetTenant(); v != "" {
			if !cmd.Flags().Changed("tenant") {
				// Set a default value to the current tenant so the user doesn't have to define one
				path.SetVariable("tenant", iterator.NewRepeatIterator(v, 1))
			}
		}
		err = flags.WithParameters(
			cmd,
			path,
			inputIterators,
			c8yfetcher.WithIDSlice(args, "tenant", "tenant"),
		)
		if err != nil {
			return err
		}

		state := h.GetStateHandler(cmd, client)

		totalErrors := 0
		var lastErr error
		var result interface{}

		for {
			tenantID, input, inputErr := path.Execute(false)

			if tenantID != "" && input == nil {
				// set the input manually when the value is not provided
				// from the pipeline, and using the default value.
				input = []byte(tenantID)
			}

			if inputErr == io.EOF {
				break
			}

			if totalErrors >= cfg.AbortOnErrorCount() {
				msg := fmt.Sprintf("Too many errors. total=%d, max=%d. lastError=%s", totalErrors, cfg.AbortOnErrorCount(), lastErr)
				return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitAbortedWithErrors, msg)
			}

			if inputErr == nil {
				// Skip checking if the input has errors
				_ = state.SetValue(tenantID)
				result, err = desiredstate.WaitForWithRetries(attempts, interval, duration, state)
				if err == nil {
					outValue := h.GetValue(result, input)
					_ = f.WriteOutputWithoutPropertyGuess(outValue, cmdutil.OutputContext{})
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
