// v2-based identity get: the name flag (the external id) drives iteration (pipe
// or --name); each external identity is fetched by type + external id via
// Identity.Get.
package get

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/identity"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get external identity
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get external identity",
		Long:  `Get an external identity object. An external identity will include the reference to a single device managed object`,
		Example: heredoc.Doc(`
$ c8y identity get --type test --name myserialnumber
Get external identity
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("type", "c8y_Serial", "External identity type")
	cmd.Flags().String("name", "", "External identity name (required) (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("name", "name", true, "externalId", "name", "id"),
		flags.WithPowershellName("Get-ExternalId"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.externalid+json", ""),
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

	if err := r.InputFlag("name"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opts := identity.IdentityOptions{
			Type:       in.String("type"),
			ExternalID: in.String("name"),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.Identity.Get(ctx, opts))
		}, nil
	})
}
