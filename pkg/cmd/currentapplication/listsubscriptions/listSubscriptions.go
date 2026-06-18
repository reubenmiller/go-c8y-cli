// v2-based current application subscriptions list: lists the subscribed users of
// the current application via Applications.Current.ListSubscriptions. The
// applicationUserCollection response embeds them under "users", so the collection
// is flattened into individual documents; the command runs once.
package listsubscriptions

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// ListSubscriptionsCmd command
type ListSubscriptionsCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListSubscriptionsCmd creates a command to Get current application subscriptions
func NewListSubscriptionsCmd(f *cmdutil.Factory) *ListSubscriptionsCmd {
	ccmd := &ListSubscriptionsCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "listSubscriptions",
		Short: "Get current application subscriptions",
		Long:  `Requires authentication with the application bootstrap user`,
		Example: heredoc.Doc(`
$ c8y currentapplication listSubscriptions
List the current application users/subscriptions
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("", "", false),
		flags.WithCollectionProperty("users"),
		flags.WithPowershellName("Get-CurrentApplicationSubscription"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.applicationUserCollection+json", "application/vnd.com.nsn.cumulocity.applicationUser+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListSubscriptionsCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		return func(ctx context.Context) output.Seq {
			return output.FromIterator(client.Applications.Current.ListSubscriptions(ctx).Items())
		}, nil
	})
}
