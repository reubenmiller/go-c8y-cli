package wait

import (
	"fmt"
	"io"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
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

	flagDurationSec int64
	expectedStatus  string
	Timeout         time.Duration
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

		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Operation id (required) (accepts pipeline)")
	cmd.Flags().StringVar(&ccmd.expectedStatus, "status", "SUCCESSFUL", "Status")
	cmd.Flags().DurationVar(&ccmd.Timeout, "duration", 30*time.Second, "Timeout. i.e. 30s or 1m (1 minute)")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
	)

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "PENDING", "EXECUTING", "SUCCESSFUL", "FAILED"),
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

	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	// path parameters
	path := flags.NewStringTemplate("{id}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		flags.WithStringValue("id", "id"),
	)
	if err != nil {
		return err
	}

	state := &c8ywaiter.OperationState{
		Client: client,
		Status: n.expectedStatus,
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
		result, err := desiredstate.WaitFor(1000*time.Millisecond, n.Timeout, state)

		if v, ok := result.(*c8y.Operation); ok {
			_ = n.factory.WriteJSONToConsole(cfg, cmd, "", []byte(v.Item.Raw))
		}

		if err != nil {
			totalErrors++
			cfg.Logger.Warnf("%s", err)
			lastErr = err
		}
	}
	return nil
}
