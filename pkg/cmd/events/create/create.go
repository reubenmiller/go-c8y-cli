// v2-based events create: the CLI builds the full request body with its body
// machinery (--data/--template/typed flags), then the device reference the body
// carries in source.id is resolved (name -> id) before the raw create call.
// This is the create bridge for source-bearing resources: the body builder is
// preserved in full; only the source reference needs go-c8y resolution.
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
)

// CreateCmd command
type CreateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateCmd creates a command to Create event
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create event",
		Long:  `Create a new event for a device`,
		Example: heredoc.Doc(`
$ c8y events create --device 12345 --type c8y_TestEvent --text "Test event"
Create a new event for a device

$ c8y events create --device MyDevice --type c8y_TestEvent --text "Test event"
Create a new event, resolving the device name to its id
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("device", "", "The ManagedObject which is the source of this event (accepts pipeline)")
	cmd.Flags().String("time", "", "Time of the event. Defaults to current timestamp")
	cmd.Flags().String("type", "", "Identifies the type of this event")
	cmd.Flags().String("text", "", "Text description of the event")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("device", "source.id", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("New-Event"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.event+json", ""),
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

	if err := r.Input(); err != nil {
		return err
	}

	err = r.Body(
		flags.WithOverrideValue("device", "source.id"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("device", "source.id"),
		flags.WithRelativeTimestamp("time", "time"),
		flags.WithStringValue("type", "type"),
		flags.WithStringValue("text", "text"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithDefaultTemplateString(`
{time: _.Now('0s')}`),
		flags.WithRequiredProperties("type", "text", "time", "source.id"),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	// go-c8y resolves the source device reference (name -> id); plain ids pass
	// through. Resolution runs against the real API even under --dry.
	resolveDevice := func(ctx context.Context, ref string) (string, error) {
		return client.Events.DeviceResolver.ResolveID(ctx, managedobjects.DeviceRef(ref), nil)
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		body, err = in.ResolveSourceID(body, resolveDevice)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Event] {
				return client.Events.CreateRaw(ctx, body)
			})
		}, nil
	})
}
