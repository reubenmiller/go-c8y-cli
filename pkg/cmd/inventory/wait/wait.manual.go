package wait

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ywait"
	assertFactory "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/assert/factory"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
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
	return assertFactory.RunWait(cmd, n.factory, "id", func(client *apiv2.Client, in *c8ystream.Resolver, ref string) c8ywait.Check {
		return c8ywait.Fragments(client, c8ystream.NameOrID(ref), n.ExpectedFragments)
	})
}
