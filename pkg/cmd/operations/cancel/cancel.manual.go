package cancel

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/desiredstate"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// UpdateCmd command
type UpdateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUpdateCmd creates a command to Update operation
func NewUpdateCmd(f *cmdutil.Factory) *UpdateCmd {
	ccmd := &UpdateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update operation",
		Long: `Update an operation. This is commonly used to change an operation's status. For example the operation can be set to FAILED along with a failure reason.
`,
		Example: heredoc.Doc(`
$ c8y operations update --id 12345 --status EXECUTING
Update an operation
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Operation id (required) (accepts pipeline)")
	cmd.Flags().String("status", "", "Operation status, can be one of SUCCESSFUL, FAILED, EXECUTING or PENDING.")
	cmd.Flags().String("failureReason", "", "Reason for the failure. Use when setting status to FAILED")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "PENDING", "EXECUTING", "SUCCESSFUL", "FAILED"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	// Runtime flag options
	flags.WithOptions(
		cmd,
		flags.WithRuntimePipelineProperty(),
	)
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"),
		flags.WithProcessingModeValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder(true)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithDataFlagValue(),
		flags.WithStringValue("status", "status"),
		flags.WithStringValue("failureReason", "failureReason"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("status"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("devicecontrol/operations/{id}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithIDSlice(args, "id", "id"),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "PUT",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}

func NewDeviceCmdFactory(cmd *cobra.Command, f *cmdutil.Factory) *cobra.Command {
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
			c8yfetcher.WithDeviceByNameFirstMatch(f, args, "device", "device"),
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
