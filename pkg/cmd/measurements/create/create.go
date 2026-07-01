// v2-based measurements create: the CLI builds the full body (measurement value
// fragments via --data/--template, plus type/time), then the device reference
// held in source.id is resolved (name -> id) before the raw create call.
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

// NewCreateCmd creates a command to Create measurement
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create measurement",
		Long:  `Create a new measurement on a device`,
		Example: heredoc.Doc(`
$ c8y measurements create --device 12345 --type c8y_Temperature --data "c8y_Temperature.T.value=25"
Create a temperature measurement on a device

$ c8y measurements create --device MyDevice --type c8y_Temperature --data "c8y_Temperature.T.value=25"
Create a measurement, resolving the device name to its id
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{}, "The ManagedObject which is the source of this measurement (accepts pipeline). Multiple devices fan out one measurement per device.")
	cmd.Flags().String("time", "", "Time of the measurement. Defaults to current timestamp")
	cmd.Flags().String("type", "", "The most specific type of this entire measurement")

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
		flags.WithPowershellName("New-Measurement"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.measurement+json", ""),
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
		flags.WithRelativeTimestamp("time", "time"),
		flags.WithStringValue("type", "type"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithDefaultTemplateString(`
{time: _.Now('0s')}`),
		flags.WithRequiredProperties("type", "time", "source.id"),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	resolveDevice := func(ctx context.Context, ref string) (string, error) {
		return client.Measurements.DeviceResolver.ResolveID(ctx, managedobjects.DeviceRef(ref), nil)
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		// Restore v1's multi-device fan-out: --device 1,2,3 zips one device per
		// piped item (cycling), overriding the source.id seeded by WithOverrideValue.
		body, err = in.ApplyOverrideSlice(body, "device", "source.id")
		if err != nil {
			return nil, err
		}
		body, err = in.ResolveSourceID(body, resolveDevice)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Measurement] {
				return client.Measurements.CreateRaw(ctx, body)
			})
		}, nil
	})
}
