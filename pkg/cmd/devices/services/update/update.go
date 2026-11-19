// v2-based update device service status: the id flag drives iteration; the
// service reference is resolved to a managed-object id and updated via
// ManagedObjects.Update. The body is built from --name/--serviceType/--status
// (+ --data/--template).
package update

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/services/serviceref"
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

// UpdateCmd command
type UpdateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUpdateCmd creates a command to Update service status
func NewUpdateCmd(f *cmdutil.Factory) *UpdateCmd {
	ccmd := &UpdateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update service status",
		Long:  `Update service status`,
		Example: heredoc.Doc(`
$ c8y devices services update --id 12345 --status up
Update service status

$ c8y devices services list --device 12345 --name ntp | c8y devices services update --status up
Update service status
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device id (required for name lookup)")
	cmd.Flags().StringSlice("id", []string{""}, "Service id (required) (accepts pipeline)")
	cmd.Flags().String("name", "", "Service name")
	cmd.Flags().String("serviceType", "", "Service type, e.g. systemd")
	cmd.Flags().String("status", "", "Service status")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDeviceService("id", "device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("status", "up", "down", "unknown"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true, "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Update-DeviceService"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObject+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("serviceType", "serviceType"),
		flags.WithStringValue("status", "status"),
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
		id, err := serviceref.ResolveID(in.ResolveContext(), client, serviceref.First(in.StringSlice("device")), in.String("id"))
		if err != nil {
			return nil, err
		}
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.ManagedObject] {
				return client.ManagedObjects.Update(ctx, id, body)
			})
		}, nil
	})
}
