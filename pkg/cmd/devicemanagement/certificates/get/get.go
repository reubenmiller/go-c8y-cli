// v2-based trusted device certificate get: the id flag drives iteration (pipe or
// --id); each reference (a fingerprint or a name, resolved to a fingerprint via
// TrustedCertificates.ResolveID) is fetched via TrustedCertificates.Get.
package get

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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get trusted device certificate
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get trusted device certificate",
		Long:  `Get a trusted device certificate`,
		Example: heredoc.Doc(`
$ c8y devicemanagement certificates get --id abcedef0123456789abcedef0123456789
Get trusted device certificate by id/fingerprint

$ c8y devicemanagement certificates get --id MyCert
Get trusted device certificate by name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Certificate fingerprint or name (accepts pipeline)")
	cmd.Flags().String("tenant", "", "Tenant id")

	completion.WithOptions(
		cmd,
		completion.WithDeviceCertificate("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", false, "fingerprint", "name", "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("Get-DeviceCertificate"),
		flags.WithOutputType("application/json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
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
		// The reference may be a fingerprint or a name; resolve to a fingerprint
		// (runs for real even under --dry so the rendered request is resolved).
		fingerprint, err := client.TrustedCertificates.ResolveID(in.ResolveContext(), tenant, in.String("id"))
		if err != nil {
			return nil, err
		}
		opt := trustedcertificates.GetOptions{TenantID: tenant, Fingerprint: fingerprint}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.TrustedCertificates.Get(ctx, opt))
		}, nil
	})
}
