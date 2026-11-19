// v2-based system options list: lists the read-only system options via
// Tenants.SystemOptions.ListAll. The endpoint has no filter parameters, so the
// typed ListOptions carries only pagination; the command runs once.
package list

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/systemoptions"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get system option collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get system option collection",
		Long:  `Get a collection of system options. This endpoint provides a set of read-only properties pre-defined in platform configuration.`,
		Example: heredoc.Doc(`
$ c8y systemoptions list
Get a list of system options
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithCollectionProperty("options"),
		flags.WithPowershellName("Get-SystemOptionCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.optionCollection+json", "application/vnd.com.nsn.cumulocity.option+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	paging, err := r.Pagination()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := systemoptions.ListOptions{}
		opt.PaginationOptions = paging
		return func(ctx context.Context) any {
			return client.Tenants.SystemOptions.ListAll(ctx, opt)
		}, nil
	})
}
