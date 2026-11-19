// v2-based identity find: the name flag (the external id) drives iteration (pipe
// or --name); each external id is looked up via the bulk search endpoint
// (POST /identity/search, Identity.Search). Unlike `get`, a missing external id
// is not an error — the search simply returns no match — so this is convenient
// for resolving a stream of external ids to their managed objects.
package find

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

// FindCmd command
type FindCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewFindCmd creates a command to Find external identities
func NewFindCmd(f *cmdutil.Factory) *FindCmd {
	ccmd := &FindCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find external identities",
		Long:  `Look up the managed objects matching one or more external identities. A missing external id yields no match rather than an error.`,
		Example: heredoc.Doc(`
$ c8y identity find --type c8y_Serial --name myserialnumber
Find an external identity by type and external id

$ c8y devices list | c8y identity find --type c8y_Serial --name "myserial*"
Resolve a stream of external ids to their managed objects
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
		flags.WithCollectionProperty("externalIds"),
		flags.WithPowershellName("Find-ExternalId"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.externalIdCollection+json", "application/vnd.com.nsn.cumulocity.externalid+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *FindCmd) RunE(cmd *cobra.Command, args []string) error {
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
		opts := identity.SearchOptions{
			ExternalIds: []identity.IdentityOptions{
				{
					Type:       in.String("type"),
					ExternalID: in.String("name"),
				},
			},
		}
		return func(ctx context.Context) output.Seq {
			return output.FromIterator(client.Identity.Search(ctx, opts).Items())
		}, nil
	})
}
