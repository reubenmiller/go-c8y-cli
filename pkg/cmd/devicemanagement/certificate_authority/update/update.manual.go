package update

import (
	"context"
	"encoding/json"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
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
			return f.UpdateModeEnabled()
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

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, cfgErr := n.factory.Config()
	if cfgErr != nil {
		return cfgErr
	}

	client, err := n.factory.Client()
	if err != nil {
		return err
	}

	certOptions := c8y.NewCertificate()

	if cmd.Flags().Changed("status") {
		certOptions.WithStatus(n.Status)
	}

	if cmd.Flags().Changed("autoRegistrationEnabled") {
		certOptions.WithAutoRegistration(n.AutoRegistrationEnabled)
	}

	cert, _, err := client.CertificateAuthority.Update(context.Background(), "", certOptions)
	if err != nil {
		return err
	}
	if cfg.DryRun() {
		return nil
	}

	b, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return n.factory.WriteOutputWithoutPropertyGuess(b, cmdutil.OutputContext{})
}
