// v2-based trusted device certificate delete: the id flag drives iteration (pipe
// or --id); each reference (a fingerprint or a name, resolved via
// TrustedCertificates.ResolveID) is removed via TrustedCertificates.Delete.
// Success yields no output.
package delete

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/trustedcertificates"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteCmd command
type DeleteCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteCmd creates a command to Delete device certificate
func NewDeleteCmd(f *cmdutil.Factory) *DeleteCmd {
	ccmd := &DeleteCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete device certificate",
		Long:  `Remove a trusted device certificate. All devices connected to the platform will no longer be able to communicate.`,
		Example: heredoc.Doc(`
$ c8y devicemanagement certificates delete --id abcedef0123456789abcedef0123456789
Remove trusted device certificate by id/fingerprint

$ c8y devicemanagement certificates delete --id MyCert
Remove trusted device certificate by name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
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
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("id", "id", false, "fingerprint", "name", "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("Remove-DeviceCertificate"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteCmd) RunE(cmd *cobra.Command, args []string) error {
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
		fingerprint, err := client.TrustedCertificates.ResolveID(in.ResolveContext(), tenant, in.String("id"))
		if err != nil {
			return nil, err
		}
		opt := trustedcertificates.DeleteOptions{TenantID: tenant, Fingerprint: fingerprint}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.TrustedCertificates.Delete(ctx, opt)
			})
		}, nil
	})
}
