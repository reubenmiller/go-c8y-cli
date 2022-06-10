package wait

import (
	"fmt"
	"io"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ywaiter"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/desiredstate"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdWait struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	ExpectedFragments []string
}

func NewCmdWait(f *cmdutil.Factory) *CmdWait {
	ccmd := &CmdWait{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Wait for managed object",
		Long:  `Wait for an managed object fragment by polling until a condition is met or a timeout is reached`,
		Example: heredoc.Doc(`
			$ c8y inventory wait --id 1234 --fragments "c8y_Mobile.iccd"
			# Wait for the managed object to have a non-null c8y_Mobile.iccd fragment

			$ c8y inventory wait --id 1234 --fragments '!c8y_Mobile'
			# Wait for the managed object to not have a c8y_Mobile fragment

			$ c8y inventory wait --id 1234 --fragments 'name=^\d+-\w+$'
			# Wait for the managed object name fragment to match the regular expression '^\d+-\w+$'

			$ c8y inventory wait --id 1234 --fragments 'name=^\d+$' --fragments c8y_IsDevice
			# Wait for the managed object name fragment to match the regular expression '^\d+$' and have the c8y_IsDevice fragment
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Inventory id (required) (accepts pipeline)")
	cmd.Flags().StringSliceVar(&ccmd.ExpectedFragments, "fragments", nil, "Fragments to wait for. If multiple values are given, then it will be applied as an OR operation")
	cmd.Flags().String("duration", "30s", "Timeout duration. i.e. 30s or 1m (1 minute)")
	cmd.Flags().String("interval", "5s", "Interval to check on the status, i.e. 10s or 1min")
	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true, "deviceId", "source.id", "managedObject.id", "id"),
	)

	completion.WithOptions(
		cmd,
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

	interval, err := flags.GetDurationFlag(cmd, "interval", true, time.Second)
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

	state := &c8ywaiter.InventoryState{
		Client:    client,
		Fragments: n.ExpectedFragments,
	}

	totalErrors := 0
	var lastErr error
	for {
		itemID, _, inputErr := path.Execute(false)

		if inputErr == io.EOF {
			break
		}

		if totalErrors >= cfg.AbortOnErrorCount() {
			msg := fmt.Sprintf("Too many errors. total=%d, max=%d. lastError=%s", totalErrors, cfg.AbortOnErrorCount(), lastErr)
			return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitAbortedWithErrors, msg)
		}

		state.ID = itemID
		result, err := desiredstate.WaitFor(interval, duration, state)

		if v, ok := result.(*c8y.ManagedObject); ok {
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
