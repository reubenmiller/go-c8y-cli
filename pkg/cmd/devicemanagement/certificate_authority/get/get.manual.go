package get

import (
	"context"
	"encoding/json"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
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
			return f.UpdateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetCmd) RunE(cmd *cobra.Command, args []string) error {
	client, err := n.factory.Client()
	if err != nil {
		return err
	}

	cert, err := client.CertificateAuthority.Get(context.Background())
	if err != nil {
		return err
	}
	if cert == nil {
		return nil
	}

	b, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return n.factory.WriteOutputWithoutPropertyGuess(b, cmdutil.OutputContext{})
}
