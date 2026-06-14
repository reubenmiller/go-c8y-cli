// v2-based alarms create: the CLI builds the full body (--data/--template/typed
// flags incl. severity/status), then the device reference held in source.id is
// resolved (name -> id) before the raw create call.
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

// NewCreateCmd creates a command to Create alarm
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create alarm",
		Long:  `Create a new alarm on a device`,
		Example: heredoc.Doc(`
$ c8y alarms create --device 12345 --type c8y_TestAlarm --text "Test alarm" --severity MAJOR
Create a major alarm on a device

$ c8y alarms create --device MyDevice --type c8y_TestAlarm --text "Test alarm" --severity CRITICAL
Create an alarm, resolving the device name to its id
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("device", "", "The ManagedObject that the alarm originated from (accepts pipeline)")
	cmd.Flags().String("type", "", "Identifies the type of this alarm, e.g. 'com_cumulocity_events_TamperEvent'")
	cmd.Flags().String("time", "", "Time of the alarm. Defaults to current timestamp")
	cmd.Flags().String("text", "", "Text description of the alarm")
	cmd.Flags().String("severity", "", "The severity of the alarm: CRITICAL, MAJOR, MINOR or WARNING")
	cmd.Flags().String("status", "", "The status of the alarm: ACTIVE, ACKNOWLEDGED or CLEARED (defaults to ACTIVE)")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("severity", "CRITICAL", "MAJOR", "MINOR", "WARNING"),
		completion.WithValidateSet("status", "ACTIVE", "ACKNOWLEDGED", "CLEARED"),
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("device", "source.id", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("New-Alarm"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.alarm+json", ""),
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
		flags.WithStringValue("severity", "severity"),
		flags.WithStringValue("status", "status"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithDefaultTemplateString(`
{time: _.Now('0s')}`),
		flags.WithRequiredProperties("type", "text", "time", "severity", "source.id"),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	resolveDevice := func(ctx context.Context, ref string) (string, error) {
		return client.Alarms.DeviceResolver.ResolveID(ctx, managedobjects.DeviceRef(ref), nil)
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
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Alarm] {
				return client.Alarms.CreateRaw(ctx, body)
			})
		}, nil
	})
}
