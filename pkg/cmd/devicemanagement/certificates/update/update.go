// v2-based trusted device certificate update: the id flag drives iteration (pipe
// or --id) and the body builder (name/status/autoRegistrationEnabled +
// --data/--template) is evaluated per item. Each reference (a fingerprint or a
// name, resolved via TrustedCertificates.ResolveID) feeds
// TrustedCertificates.Update.
package update

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/trustedcertificates"
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

// NewUpdateCmd creates a command to Update trusted device certificate
func NewUpdateCmd(f *cmdutil.Factory) *UpdateCmd {
	ccmd := &UpdateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update trusted device certificate",
		Long:  `Update settings of an existing trusted device certificate`,
		Example: heredoc.Doc(`
$ c8y devicemanagement certificates update --id abcedef0123456789abcedef0123456789 --status DISABLED
Update device certificate by id/fingerprint

$ c8y devicemanagement certificates update --id MyCert --status DISABLED
Update device certificate by name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Certificate fingerprint or name (accepts pipeline)")
	cmd.Flags().String("tenant", "", "Tenant id")
	cmd.Flags().String("name", "", "Certificate name")
	cmd.Flags().String("status", "", "Status")
	cmd.Flags().Bool("autoRegistrationEnabled", false, "Enable auto registration")

	completion.WithOptions(
		cmd,
		completion.WithDeviceCertificate("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("status", "ENABLED", "DISABLED"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", false, "fingerprint", "name", "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("Update-DeviceCertificate"),
		flags.WithOutputType("application/json", ""),
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
		flags.WithStringValue("status", "status"),
		flags.WithBoolValue("autoRegistrationEnabled", "autoRegistrationEnabled", ""),
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
		tenant := in.String("tenant")
		if tenant == "" {
			tenant = n.factory.GetTenant()
		}
		fingerprint, err := client.TrustedCertificates.ResolveID(in.ResolveContext(), tenant, in.String("id"))
		if err != nil {
			return nil, err
		}
		opt := trustedcertificates.UpdateOptions{TenantID: tenant, Fingerprint: fingerprint}
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.TrustedCertificate] {
				return client.TrustedCertificates.Update(ctx, opt, body)
			})
		}, nil
	})
}
