package certificate_authority

import (
	"context"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/trustedcertificates/certificateauthority"
	"github.com/spf13/cobra"
)

// DeleteCmd command
type DeleteCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	Status                  string
	AutoRegistrationEnabled bool
}

// NewDeleteCmd creates a command to delete the tenant's certificate authority
func NewDeleteCmd(f *cmdutil.Factory) *DeleteCmd {
	ccmd := &DeleteCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete tenant certificate authority",
		Long:  "Delete tenant certificate authority",
		Example: heredoc.Doc(`
$ c8y devicemanagement certificate-authority delete
Delete certificate authority
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	flags.WithOptions(
		cmd,

		// Enable confirmation prompts
		flags.WithSemanticMethod("DELETE"),
	)

	cmd.SilenceUsage = true

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	client, err := r.Client()
	if err != nil {
		return err
	}
	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		return func(ctx context.Context) any {
			return client.TrustedCertificates.CertificateAuthority.Delete(ctx, certificateauthority.DeleteOptions{})
		}, nil
	})
}
