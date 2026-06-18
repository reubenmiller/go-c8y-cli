// v2-based trusted device certificate create: builds the body (name/status/
// autoRegistrationEnabled + the certificate file read into certInPemFormat, plus
// --data/--template) and uploads it via TrustedCertificates.Create. The name
// flag drives iteration (pipe or --name). The SDK Create takes the tenant via
// CreateOptions, defaulting to the current tenant.
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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/trustedcertificates"
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

// NewCreateCmd creates a command to Upload trusted device certificate
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Upload trusted device certificate",
		Long:  `Upload a trusted device certificate which will enable communication to Cumulocity using the certificate (or a cert which is trusted by the certificate)`,
		Example: heredoc.Doc(`
$ c8y devicemanagement certificates create --name "MyCert" --file "trustedcert.pem"
Upload a trusted device certificate

$ c8y devicemanagement certificates list | c8y devicemanagement certificates create --template input.value --session c8y.Q.instance
Copy device certificates from one Cumulocity tenant to another (tenants must not be hosted on the same instance!)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant id")
	cmd.Flags().String("name", "", "Certificate name (accepts pipeline)")
	cmd.Flags().String("status", "ENABLED", "Status")
	cmd.Flags().String("file", "", "Certificate file (in PEM format with header/footer)")
	cmd.Flags().Bool("autoRegistrationEnabled", false, "Enable auto registration")

	completion.WithOptions(
		cmd,
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("status", "ENABLED", "DISABLED"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("name", "name", false, "name"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("New-DeviceCertificate"),
		flags.WithOutputType("application/json", ""),
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
		flags.WithOverrideValue("name", "name"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("status", "status"),
		flags.WithCertificateFile("file", "certInPemFormat"),
		flags.WithBoolValue("autoRegistrationEnabled", "autoRegistrationEnabled", ""),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("name", "certInPemFormat", "status"),
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
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		opt := trustedcertificates.CreateOptions{TenantID: tenant}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.TrustedCertificate] {
				return client.TrustedCertificates.Create(ctx, opt, body)
			})
		}, nil
	})
}
