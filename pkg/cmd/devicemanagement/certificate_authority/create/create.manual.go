package create

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

// CreateCmd command
type CreateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	Status                  string
	AutoRegistrationEnabled bool
}

// NewDeleteCmd creates a command to create the tenant's certificate authority
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create tenant certificate authority",
		Long: `Create a key pair and self-sign a certificate with as the Common Name (CN).
Store the private key in an encrypted tenant option.

The devices can be registered automatically only when device administrator checks
this option ON.

If the CA certificate is removed from the trusted certificate list,
corresponding public and private key removed automatically from the database collection.
If a CA is already present, return a message indicating the CA is already present.
`,
		Example: heredoc.Doc(`
$ c8y devicemanagement certificate-authority create
Create new certificate authority (and enable auto registration)

$ c8y devicemanagement certificate-authority create --status DISABLED
Create new certificate authority but disable it
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().BoolVar(&ccmd.AutoRegistrationEnabled, "autoRegistrationEnabled", true, "Enable auto registration")
	cmd.Flags().StringVar(&ccmd.Status, "status", "", "Status")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "ENABLED", "DISABLED"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, cfgErr := n.factory.Config()
	if cfgErr != nil {
		return cfgErr
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}

	cert, err := client.CertificateAuthority.Create(context.Background(), c8y.CertificateAuthorityOptions{
		AutoRegistration: n.AutoRegistrationEnabled,
		Status:           n.Status,
	})
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
