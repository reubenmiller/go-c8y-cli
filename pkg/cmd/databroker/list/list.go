// v2-based data-broker connector list: lists the data-broker connectors via
// DataBroker.ListAll.
package list

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/databroker"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get data broker collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get data broker collection",
		Long:  `Get a collection of existing data broker connectors`,
		Example: heredoc.Doc(`
$ c8y databroker list
Get a list of data broker connectors
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithCollectionProperty("connectors"),
		flags.WithPowershellName("Get-DataBrokerConnectorCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.databrokerConnectorCollection+json", "application/vnd.com.nsn.cumulocity.databrokerConnector+json"),
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
		opt := databroker.ListOptions{}
		opt.PaginationOptions = paging
		return func(ctx context.Context) any {
			return client.DataBroker.ListAll(ctx, opt)
		}, nil
	})
}
