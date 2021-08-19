package assert

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ywaiter"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/desiredstate"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdAssert struct {
	*subcommand.SubCommand

	negate  bool
	exists  bool
	retries int64

	factory *cmdutil.Factory
}

func NewCmdAssert(f *cmdutil.Factory) *CmdAssert {
	ccmd := &CmdAssert{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "assert",
		Short: "Assert existance of a managed object",
		Long: heredoc.Doc(`
			Assert that a managed objects exists or not and pass input untouched

			If the assertion is true, then the input value (stdin or an explicit argument value) will be passed untouched to stdout.
			This is useful if you want to filter a list of managed objects by whether they exist or not in the platform, and use the results
			in some downstream command (in the pipeline)
		`),
		Example: heredoc.Doc(`
			$ c8y inventory assert --exists --id 1234
			# Assert the managed object exists

			$ echo "1111" | c8y inventory assert --exists
			# Pass the piped input only on if the ids exists as a managed object

			$ echo -e "1111\n2222" | c8y inventory assert --not --exists
			# Only select the managed object ids which do not exist
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Inventory id (required) (accepts pipeline)")
	cmd.Flags().String("duration", "30s", "Timeout duration. i.e. 30s or 1m (1 minute)")
	cmd.Flags().String("interval", "5s", "Interval to check on the status, i.e. 10s or 1min")
	cmd.Flags().Int64Var(&ccmd.retries, "retries", 0, "Interval to check on the status, i.e. 10s or 1min")
	cmd.Flags().BoolVar(&ccmd.negate, "not", false, "Negate the match")
	cmd.Flags().BoolVar(&ccmd.exists, "exists", true, "Assert the existance")
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

func (n *CmdAssert) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}

	consol, err := n.factory.Console()
	if err != nil {
		return err
	}

	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	interval, err := flags.GetDurationFlag(cmd, "interval", true, time.Second)
	if err != nil {
		return err
	}

	timeout, err := flags.GetDurationFlag(cmd, "timeout", true, time.Second)
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

	state := &c8ywaiter.InventoryExistance{
		Client: client,
		Negate: n.negate,
	}

	totalErrors := 0
	var lastErr error
	for {
		itemID, input, inputErr := path.Execute(false)

		if inputErr == io.EOF {
			break
		}

		if totalErrors >= cfg.AbortOnErrorCount() {
			msg := fmt.Sprintf("Too many errors. total=%d, max=%d. lastError=%s", totalErrors, cfg.AbortOnErrorCount(), lastErr)
			return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitAbortedWithErrors, msg)
		}

		state.ID = itemID
		result, err := desiredstate.WaitForWithRetries(n.retries, interval, timeout, state)

		if v, ok := result.(*c8y.ManagedObject); ok {
			_ = n.factory.WriteJSONToConsole(cfg, cmd, "", []byte(v.Item.Raw))
		} else {
			if err == nil {
				switch v := input.(type) {
				case []byte:
					if jsonUtilities.IsJSONObject(v) {
						_ = n.factory.WriteJSONToConsole(cfg, cmd, "", v)
					} else {
						fmt.Fprintf(consol, "%s\n", v)
					}
				}
			}
		}

		if err != nil && !strings.Contains(err.Error(), "Max retries exceeded") {
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
