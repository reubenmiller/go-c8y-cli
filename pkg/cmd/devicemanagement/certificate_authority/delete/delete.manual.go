package delete

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/worker"
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
	cfg, cfgErr := n.factory.Config()
	if cfgErr != nil {
		return cfgErr
	}

	cs := n.factory.IOStreams.ColorScheme()
	client, err := n.factory.Client()
	if err != nil {
		return err
	}

	return n.factory.RunWithGenericWorkers(cmd, nil, nil, func(j worker.Job) (any, error) {
		_, err = client.CertificateAuthority.Delete(context.Background(), "")
		if err != nil {
			return nil, err
		}
		if cfg.DryRun() {
			return nil, nil
		}
		_, err = fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s Deleted certificate-authority for tenant %s\n", cs.SuccessIconWithColor(cs.Red), client.GetTenantName(context.Background()))
		return nil, err
	})
}
