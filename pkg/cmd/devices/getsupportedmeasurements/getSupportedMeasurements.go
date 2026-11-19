// v2-based get supported measurements: the device flag drives iteration; each
// device is resolved (name -> id) and its supported measurement fragments are
// fetched via Devices.ListSupportedMeasurements and rendered as-is (the spec's
// collectionProperty "-").
package getsupportedmeasurements

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetSupportedMeasurementsCmd command
type GetSupportedMeasurementsCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetSupportedMeasurementsCmd creates a command to Get supported measurements
func NewGetSupportedMeasurementsCmd(f *cmdutil.Factory) *GetSupportedMeasurementsCmd {
	ccmd := &GetSupportedMeasurementsCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "getSupportedMeasurements",
		Short: "Get supported measurements",
		Long: `Returns a list of fragments (valueFragmentTypes) related to the device
`,
		Example: heredoc.Doc(`
$ c8y devices getSupportedMeasurements --device 12345
Get the supported measurements of a device by name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device ID (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "device", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("-"),
		flags.WithPowershellName("Get-SupportedMeasurements"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.inventory+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetSupportedMeasurementsCmd) RunE(cmd *cobra.Command, args []string) error {
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

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		id, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("device")), nil)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.Devices.ListSupportedMeasurements(ctx, id))
		}, nil
	})
}
