// v2-based retention rules list: retention rules expose no filter parameters, so
// the typed ListOptions carries only pagination. There is no pipeline input, so
// the command runs exactly once.

package retentionrules

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/retentionrules"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get retention rule collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get retention rule collection",
		Long:  `Get a collection of retention rules configured in the current tenant`,
		Example: heredoc.Doc(`
$ c8y retentionrules list
Get a list of retention rules
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithCollectionProperty("retentionRules"),
		flags.WithPowershellName("Get-RetentionRuleCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.retentionRuleCollection+json", "application/vnd.com.nsn.cumulocity.retentionRule+json"),
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
		opt := retentionrules.ListOptions{}
		opt.PaginationOptions = paging
		return func(ctx context.Context) any {
			return client.RetentionRules.ListAll(ctx, opt)
		}, nil
	})
}
