package update

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/trustedcertificates/certificateauthority"
	"github.com/spf13/cobra"
)

// UpdateCmd command
type UpdateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	Status                  string
	AutoRegistrationEnabled bool
}

// NewUpdateCmd creates a command to (PREVIEW FEATURE) Create tenant certificate authority
func NewUpdateCmd(f *cmdutil.Factory) *UpdateCmd {
	ccmd := &UpdateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update tenant certificate authority",
		Long:  "Update tenant certificate authority",
		Example: heredoc.Doc(`
$ c8y devicemanagement certificate-authority update --autoRegistrationEnabled
Update certificate authority to allow auto registration

$ c8y devicemanagement certificate-authority update --status DISABLED
Disable certificate authority
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().BoolVar(&ccmd.AutoRegistrationEnabled, "autoRegistrationEnabled", false, "Enable auto registration")
	cmd.Flags().StringVar(&ccmd.Status, "status", "", "Status. Can be either ENABLED or DISABLED")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "ENABLED", "DISABLED"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),

		// Enable confirmation prompts
		flags.WithSemanticMethod("PUT"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("status", "status"),
		flags.WithBoolValue("autoRegistrationEnabled", "autoRegistrationEnabled"),
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
		return func(ctx context.Context) any {
			return client.TrustedCertificates.CertificateAuthority.Update(ctx, certificateauthority.UpdateOptions{}, body)
		}, nil
	})
}
