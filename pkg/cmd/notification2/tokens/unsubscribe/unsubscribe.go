// v2-based notification2 unsubscribe: the token drives iteration and is sent as
// a query parameter; the {result} response is rendered as-is.
package unsubscribe

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/notification2"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// UnsubscribeCmd command
type UnsubscribeCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUnsubscribeCmd creates a command to Unsubscribe via a token
func NewUnsubscribeCmd(f *cmdutil.Factory) *UnsubscribeCmd {
	ccmd := &UnsubscribeCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "unsubscribe",
		Short: "Unsubscribe via a token",
		Long: `Unsubscribe a notification subscriber using the notification token
Once a subscription is made, notifications will be kept until they are consumed by all subscribers who have previously connected to the subscription.

For non-volatile subscriptions, this can result in notifications remaining in storage if never consumed by the application.
They will be deleted if a tenant is deleted. It can take up considerable space in permanent storage for high-frequency notification sources.
Therefore, we recommend you to unsubscribe a subscriber that will never run again.
`,
		Example: heredoc.Doc(`
$ c8y notification2 tokens unsubscribe --token "eyJhbGciOiJSUzI1NiJ9"
Unsubscribe a subscriber using its token
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("token", "", "Subscriptions associated with this token will be removed (required) (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("token", "token", true, "id"),
		flags.WithPowershellName("Unregister-Notification2Subscriber"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UnsubscribeCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("token"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		token := in.String("token")
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitValue(ctx, apiv2.IsDryRun(ctx), func(ctx context.Context) op.Result[notification2.UnsubscribeResponse] {
				return client.Notification2.UnsubscribeSubscriber(ctx, token)
			})
		}, nil
	})
}
