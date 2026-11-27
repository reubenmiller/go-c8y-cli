// v2-based datahub users list: lists the DataHub (Dremio) users via
// DataHub.ListUsers. The endpoint returns one document whose data.users array
// is flattened into individual items. Runs once (no pipeline).

package users

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

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to List the data hub users
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the data hub users",
		Long:  `List the data hub users`,
		Example: heredoc.Doc(`
$ c8y datahub users list
List the datahub users
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithCollectionProperty("data.users"),
		flags.WithPowershellName("Get-DataHubUserCollection"),
		flags.WithOutputType("application/json", ""),
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
	// the Data Hub api does not take the page size of the Cumulocity collections
	r.NoDefaultPageSize()

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		return func(ctx context.Context) any {
			return output.FromIterator(client.DataHub.ListUsers(ctx).Items())
		}, nil
	})
}
