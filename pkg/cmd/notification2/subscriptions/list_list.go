// v2-based notification2 subscriptions list: fills the typed
// notification2.ListOptions; the device reference is resolved (name -> id) by
// go-c8y into the source query parameter. The device flag drives iteration.

package subscriptions

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/notification2"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get subscription collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get subscription collection",
		Long:  `Retrieve all subscriptions on your tenant, or a specific subset based on queries.`,
		Example: heredoc.Doc(`
$ c8y notification2 subscriptions list
Get existing subscriptions

$ c8y notification2 subscriptions list --context mo
Get all subscriptions for the managed object scope

$ c8y notification2 subscriptions list --device 12345
Get all subscriptions related to a specific source
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "The managed object ID to which the subscription is associated. (accepts pipeline)")
	cmd.Flags().String("context", "", "The context to which the subscription is associated.")
	cmd.Flags().String("subscription", "", "The subscription name by which filtering will be done. >= 1016.x")
	cmd.Flags().String("typeFilter", "", "The type used to filter subscriptions. This will check the subscription's subscriptionFilter.typeFilter field. >= 1016.x")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", ccmd.factory.ClientV2),
		completion.WithValidateSet("context", "mo", "tenant"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "source", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("subscriptions"),
		flags.WithPowershellName("Get-Notification2SubscriptionCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.subscriptioncollection+json", "application/vnd.com.nsn.cumulocity.subscription+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("device"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	paging, err := r.Pagination()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := notification2.ListOptions{
			Context:      in.String("context"),
			Subscription: in.String("subscription"),
			TypeFilter:   in.String("typeFilter"),
		}
		if device := in.String("device"); device != "" {
			opt.Source = c8ystream.NameOrID(device)
		}
		opt.PaginationOptions = paging
		return func(ctx context.Context) any {
			return client.Notification2.ListAll(ctx, opt)
		}, nil
	})
}
