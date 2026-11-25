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

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	Status                  string
	AutoRegistrationEnabled bool
}

// NewGetCmd creates a command to get the tenant's certificate authority
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get tenant certificate authority",
		Long:  "Get tenant certificate authority",
		Example: heredoc.Doc(`
$ c8y devicemanagement certificate-authority get
Get certificate authority to allow auto registration
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	flags.WithOptions(
		cmd,

		// Enable confirmation prompts
		flags.WithSemanticMethod("GET"),
	)

	cmd.SilenceUsage = true

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetCmd) RunE(cmd *cobra.Command, args []string) error {
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
			return client.TrustedCertificates.CertificateAuthority.Get(ctx, certificateauthority.GetOptions{})
		}, nil
	})
}
