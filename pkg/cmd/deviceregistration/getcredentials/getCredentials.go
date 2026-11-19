// v2-based device registration getCredentials: requests the credentials for a
// device that does not yet have any. The id flag drives iteration (pipe or --id);
// its value is carried on the body and posted to the device-credentials endpoint
// via Devices.Registration.CreateCredentialsRaw. It is semantically a read, so it
// is not subject to a confirmation prompt.
package getcredentials

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

// GetCredentialsCmd command
type GetCredentialsCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCredentialsCmd creates a command to Request device credentials
func NewGetCredentialsCmd(f *cmdutil.Factory) *GetCredentialsCmd {
	ccmd := &GetCredentialsCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "getCredentials",
		Short: "Request device credentials",
		Long:  `Device credentials can be enquired by devices that do not have credentials for accessing a tenant yet. Since the device does not have credentials yet, a set of fixed credentials is used for this API. The credentials can be obtained by contacting support. Do not use your tenant credentials with this API.`,
		Example: heredoc.Doc(`
$ c8y deviceregistration getCredentials --id "device-AD76-matrixer"
Request credentials for a new device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Device identifier. Max: 1000 characters. E.g. IMEI (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithDeviceRegistrationRequest("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("Request-DeviceCredentials"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.deviceCredentials+json", ""),
		flags.WithSemanticMethod("GET"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetCredentialsCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
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
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		// The id flag is the iterating driver (a string slice), so carry its
		// resolved value onto the body here rather than via a body getter.
		if body, err = sjson.SetBytes(body, "id", in.String("id")); err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.DeviceCredentials] {
				return client.Devices.Registration.CreateCredentialsRaw(ctx, body)
			})
		}, nil
	})
}
