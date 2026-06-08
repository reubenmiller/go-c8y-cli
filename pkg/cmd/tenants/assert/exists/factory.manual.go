package exists

import (
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/desiredstate"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/worker"
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

		return f.RunWithGenericWorkers(cmd, inputIterators, path, func(j worker.Job) (any, error) {
			state := h.GetStateHandler(cmd, client)

			itemID := string(j.Value.([]byte))
			input := j.Input

			// when the value is not provided from the pipeline (e.g. using the
			// default value or an explicit flag), use the resolved id as the input
			// so it can be passed untouched to the output
			if _, ok := input.([]byte); !ok {
				input = j.Value
			}

			// Skip checking if the input has errors
			_ = state.SetValue(itemID)
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
