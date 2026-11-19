// v2-based system options get: the key flag drives iteration (pipe or --key);
// each system option is fetched by category + key via
// Tenants.SystemOptions.Get. System options are read-only platform properties.
package get

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/systemoptions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get system option
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get system option",
		Long:  `Get a system option by category and key`,
		Example: heredoc.Doc(`
$ c8y systemoptions get --category "system" --key "version"
Get a system option by category and key
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("category", "", "System Option category (required)")
	cmd.Flags().String("key", "", "System Option key (required) (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("key", "key", true, "id"),
		flags.WithPowershellName("Get-SystemOption"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.option+json", ""),
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

	if err := r.InputFlag("key"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := systemoptions.GetOption{
			Category: in.String("category"),
			Key:      in.String("key"),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.Tenants.SystemOptions.Get(ctx, opt))
		}, nil
	})
}
