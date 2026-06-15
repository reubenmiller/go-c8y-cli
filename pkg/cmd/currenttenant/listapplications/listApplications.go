// v2-based current-tenant application list: lists the applications subscribed to
// the current tenant via Tenants.Current.ListApplications. The current-tenant
// endpoint embeds them, so the collection is flattened into individual
// documents; the command runs once.
package listapplications

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// ListApplicationsCmd command
type ListApplicationsCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListApplicationsCmd creates a command to List applications in current tenant
func NewListApplicationsCmd(f *cmdutil.Factory) *ListApplicationsCmd {
	ccmd := &ListApplicationsCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "listApplications",
		Short: "List applications in current tenant",
		Long:  `Get the applications of the current tenant`,
		Example: heredoc.Doc(`
$ c8y currenttenant listApplications
Get a list of applications in the current tenant
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithCollectionProperty("applications.references.#.application"),
		flags.WithPowershellName("Get-CurrentTenantApplicationCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.currentTenant+json", "application/vnd.com.nsn.cumulocity.application+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListApplicationsCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		return func(ctx context.Context) output.Seq {
			return output.FromIterator(client.Tenants.Current.ListApplications(ctx).Items())
		}, nil
	})
}
