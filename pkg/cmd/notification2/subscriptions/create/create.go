// v2-based notification2 subscription create: the CLI builds the body
// (subscription/context, the subscriptionFilter apis/typeFilter via a jsonnet
// template, plus --data/--template), the device flag drives iteration and its
// value is resolved (name -> id) into the body's source.id before the raw
// create call — so the same subscription can be created for many piped devices.
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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
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

// NewCreateCmd creates a command to Create subscription
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create subscription",
		Long:  `Create a subscription`,
		Example: heredoc.Doc(`
$ c8y notification2 subscriptions create --name deviceSub --device 12345 --context mo --apiFilter operations --apiFilter alarms
Create a new subscription to operations for a specific device

$ echo -e "1111\n2222" | c8y notification2 subscriptions create --name devicegroup --context mo --apiFilter operations
Create a subscription which groups all devices in a single subscription name

$ c8y devices list | c8y notification2 subscriptions create --name devicegroup --context mo --apiFilter operations
Create a subscription which groups all devices in a single subscription name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "The managed object to which the subscription is associated. (accepts pipeline)")
	cmd.Flags().String("name", "", "The subscription name. Each subscription is identified by a unique name within a specific context.")
	cmd.Flags().String("context", "", "The context to which the subscription is associated.")
	cmd.Flags().StringSlice("fragmentsToCopy", []string{""}, "Transforms the data to only include specified custom fragments. Each custom fragment is identified by a unique name. If nothing is specified here, the data is forwarded as-is.")
	cmd.Flags().StringSlice("apiFilter", []string{""}, "Filter notifications by api")
	cmd.Flags().StringSlice("typeFilter", []string{""}, "The data needs to have the specified value in its type property to meet the filter criteria.")
	cmd.Flags().Bool("nonPersistent", false, "Indicates whether the messages for this subscription are persistent or non-persistent, meaning they can be lost if consumer is not connected. >= 1016.x")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithNotification2SubscriptionName("name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("context", "mo", "tenant"),
		completion.WithValidateSet("apiFilter", "alarms", "alarmsWithChildren", "events", "eventsWithChildren", "managedobjects", "measurements", "operations", "*"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("device", "source.id", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("New-Notification2Subscription"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.subscriptioncollection+json", ""),
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

	// The device flag drives iteration: piped device objects feed it (id /
	// source.id extracted), or its own --device value drives a single run, so the
	// same subscription can be created for many devices from a pipe.
	if err := r.InputFlag("device"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "subscription"),
		flags.WithStringValue("context", "context"),
		flags.WithStringSliceValues("fragmentsToCopy", "fragmentsToCopy", ""),
		flags.WithStringSliceValues("apiFilter", "_apis", ""),
		flags.WithStringSliceValues("typeFilter", "_typeFilters", "'%s'"),
		flags.WithBoolValue("nonPersistent", "nonPersistent", ""),
		flags.WithRequiredTemplateString(`
{
  'subscriptionFilter': {
    [if std.length($._typeFilters) > 0 then 'typeFilter']: std.join(' or ', $._typeFilters),
    [if std.length($._apis) > 0 then 'apis']: $._apis,
  },
  _typeFilters:: [],
  _apis:: [],
}
`),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("context", "subscription"),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	resolveDevice := func(ctx context.Context, ref string) (string, error) {
		return client.Notification2.DeviceResolver.ResolveID(ctx, managedobjects.DeviceRef(ref), nil)
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		// Carry the (resolved) source id on the subscription. The device value
		// comes from the driver so a piped device populates it; ResolveSourceID
		// applies the name-or-id convention and resolves the name against the real
		// API even under --dry. No device -> no source (e.g. tenant context).
		if device := in.String("device"); device != "" {
			if body, err = sjson.SetBytes(body, "source.id", device); err != nil {
				return nil, err
			}
		}
		if body, err = in.ResolveSourceID(body, resolveDevice); err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Notification2Subscription] {
				return client.Notification2.CreateRaw(ctx, body)
			})
		}, nil
	})
}
