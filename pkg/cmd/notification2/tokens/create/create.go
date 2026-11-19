// v2-based notification2 token create: the CLI builds the token body
// (subscription/expiry/flags, plus --data/--template), the subscriber flag
// drives iteration and its value (or the 'goc8ycli' default) is written onto the
// body before the raw create call.
package create

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
	"github.com/tidwall/sjson"
)

// CreateCmd command
type CreateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateCmd creates a command to Create a token
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a token",
		Long:  `Create a token to use for subscribing to notifications`,
		Example: heredoc.Doc(`
$ c8y notification2 tokens create --name testSubscription --subscriber testSubscriber --expiresInMinutes 1440
Create a new token for a subscription which is valid for 1 day

$ c8y notification2 tokens create --name testSubscription --subscriber testSubscriber --expiresInMinutes 30
Create a new token which is valid for 30 minutes
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("subscriber", "", "The subscriber name which the client wishes to be identified with. (accepts pipeline)")
	cmd.Flags().String("name", "", "The subscription name. This value must match the same that was used when the subscription was created.")
	cmd.Flags().Int("expiresInMinutes", 1440, "The token expiration duration.")
	cmd.Flags().Bool("shared", false, "Subscription is shared amongst multiple subscribers. >= 1016.x")
	cmd.Flags().String("type", "", "The subscription type. Currently the only supported type is notification .Other types may be added in future.")
	cmd.Flags().Bool("signed", false, "If true, the token will be securely signed by the Cumulocity platform. >= 1016.x")
	cmd.Flags().Bool("nonPersistent", false, "If true, indicates that the created token refers to the non-persistent variant of the named subscription. >= 1016.x")

	completion.WithOptions(
		cmd,
		completion.WithNotification2SubscriptionName("name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("type", "notification"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("subscriber", "subscriber", false, "id"),
		flags.WithPowershellName("New-Notification2Token"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	// The subscriber flag drives iteration so a single subscription can be
	// tokenised for several piped subscribers.
	if err := r.InputFlag("subscriber"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "subscription"),
		flags.WithIntValue("expiresInMinutes", "expiresInMinutes"),
		flags.WithBoolValue("shared", "shared", ""),
		flags.WithStringValue("type", "type"),
		flags.WithBoolValue("signed", "signed", ""),
		flags.WithBoolValue("nonPersistent", "nonPersistent", ""),
		flags.WithDefaultTemplateString(`
{subscriber: 'goc8ycli'}
`),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("subscriber", "subscription"),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		// The subscriber comes from the driver, so a piped subscriber populates the
		// body (where a bound getter would not, per the piped-create convention);
		// an unset subscriber keeps the 'goc8ycli' default from the template.
		if subscriber := in.String("subscriber"); subscriber != "" {
			if body, err = sjson.SetBytes(body, "subscriber", subscriber); err != nil {
				return nil, err
			}
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Notification2Token] {
				return client.Notification2.CreateTokenRaw(ctx, body)
			})
		}, nil
	})
}
