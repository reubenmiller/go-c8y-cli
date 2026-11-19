// v2-based device registration approve: the id flag drives iteration (pipe or
// --id) and the body (status, defaulting to ACCEPTED, plus an optional security
// token) is evaluated per item. Each device request id (the device's external
// id, used as-is) is updated via Devices.Registration.UpdateRaw.
package approve

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

// ApproveCmd command
type ApproveCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewApproveCmd creates a command to Approve device request
func NewApproveCmd(f *cmdutil.Factory) *ApproveCmd {
	ccmd := &ApproveCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve device request",
		Long:  `Approve a new device request. Note: a device can only be approved if the platform has received a request for device credentials.`,
		Example: heredoc.Doc(`
$ c8y deviceregistration approve --id "1234010101s01ldk208"
Approve a new device request

$ c8y deviceregistration approve --id "1234010101s01ldk208" --securityToken "abcdef123456"
Approve a new device request and provide a security token
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Device identifier (required) (accepts pipeline)")
	cmd.Flags().String("status", "", "Status of registration")
	cmd.Flags().String("securityToken", "", "When accepting a device request, the security token is verified against the token submitted by the device when requesting credentials")

	completion.WithOptions(
		cmd,
		completion.WithDeviceRegistrationRequest("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("status", "ACCEPTED"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("Approve-DeviceRequest"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.newDeviceRequest+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ApproveCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("status", "status"),
		flags.WithStringValue("securityToken", "securityToken"),
		flags.WithDefaultTemplateString(`
{status: 'ACCEPTED'}`),
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
		id := in.String("id")
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.DeviceRequest] {
				return client.Devices.Registration.UpdateRaw(ctx, id, body)
			})
		}, nil
	})
}
