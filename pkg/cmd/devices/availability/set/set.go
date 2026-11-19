// v2-based set required availability: the id flag drives iteration; each device's
// c8y_RequiredAvailability.responseInterval is set via a managed-object update
// (Devices.Update, which resolves the name -> id). The body is built from
// --interval (+ --data/--template).
package set

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
)

// SetCmd command
type SetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewSetCmd creates a command to Set required availability
func NewSetCmd(f *cmdutil.Factory) *SetCmd {
	ccmd := &SetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set required availability",
		Long:  `Set the required availability of a device. Devices that have not sent any message in the response interval are considered unavailable. Response interval can have value between -32768 and 32767 and any values out of range will be shrink to range borders. Such devices are marked as unavailable (see below) and an unavailability alarm is raised. Devices with a response interval of zero minutes are considered to be under maintenance. No alarm is raised while a device is under maintenance. Devices that do not contain 'c8y_RequiredAvailability' are not monitored.`,
		Example: heredoc.Doc(`
$ c8y devices availability set --id 12345 --interval 10
Set the required availability of a device by name to 10 minutes

$ c8y devices get --id device01 --dry=false | c8y devices availability set --interval 10
Set the required availability for a device using pipeline

$ c8y devices list | c8y devices availability set --interval 10
Set the required availability for a list of devices using pipeline
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Device ID (required) (accepts pipeline)")
	cmd.Flags().Int("interval", 0, "Interval in minutes (required)")

	completion.WithOptions(
		cmd,
		completion.WithDevice("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("id", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Set-DeviceRequiredAvailability"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.inventory+json", ""),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("interval")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *SetCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithIntValue("interval", "c8y_RequiredAvailability.responseInterval"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		ref := c8ystream.NameOrID(in.String("id"))
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.ManagedObject] {
				return client.Devices.Update(ctx, ref, body)
			})
		}, nil
	})
}
