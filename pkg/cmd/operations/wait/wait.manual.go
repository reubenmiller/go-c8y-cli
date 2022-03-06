package wait

import (
	"fmt"
	"io"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ywaiter"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/desiredstate"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdWait struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	ExpectedStatus []string
}

func NewCmdWait(f *cmdutil.Factory) *CmdWait {
	ccmd := &CmdWait{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Wait for operation",
		Long:  `Wait for an operation to complete`,
		Example: heredoc.Doc(`
			$ c8y operations wait --id 1234
			# Wait for the operation to be set to SUCCESSFUL

			$ c8y operations wait --id 1234 --duration 1m
			# Wait for the operation to be set to SUCCESSFUL and give up after 1 minute

			$ c8y operations list --device 1111 | c8y operations wait --status "FAILED" --status "SUCCESSFUL"
			# Wait for operation to be set to either FAILED or SUCCESSFUL
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Operation id (required) (accepts pipeline)")
	cmd.Flags().StringSliceVar(&ccmd.ExpectedStatus, "status", []string{"SUCCESSFUL"}, "Status to wait for. If multiple values are given, then it will be applied as an OR operation")
	cmd.Flags().String("duration", "30s", "Timeout duration. i.e. 30s or 1m (1 minute)")
	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
	)

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", c8y.OperationStatusPending, c8y.OperationStatusExecuting, c8y.OperationStatusSuccessful, c8y.OperationStatusFailed),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdWait) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
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

	state := &c8ywaiter.OperationState{
		Client: client,
		Status: n.ExpectedStatus,
	}

	totalErrors := 0
	var lastErr error
	for {
		operationID, _, inputErr := path.Execute(false)

		if inputErr == io.EOF {
			break
		}

		if totalErrors >= cfg.AbortOnErrorCount() {
			msg := fmt.Sprintf("Too many errors. total=%d, max=%d. lastError=%s", totalErrors, cfg.AbortOnErrorCount(), lastErr)
			return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitAbortedWithErrors, msg)
		}

		state.ID = operationID
		result, err := desiredstate.WaitFor(1000*time.Millisecond, duration, state)

		if v, ok := result.(*c8y.Operation); ok {
			if v.FailureReason != "" {
				cfg.Logger.Warnf("Failure reason: %s", v.FailureReason)
			}
			_ = n.factory.WriteJSONToConsole(cfg, cmd, "", []byte(v.Item.Raw))
		}

		if err != nil {
			totalErrors++
			cfg.Logger.Infof("%s", err)
			lastErr = err
		}
	}
	if totalErrors > 0 {
		return lastErr
	}
	return nil
}
