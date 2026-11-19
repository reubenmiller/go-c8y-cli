// v2-based datahub configuration list: lists the DataHub offloading
// configurations via DataHub.ListOffloadingConfigurations. The management API
// pages with "limit", so the CLI's pageSize maps onto it (the v1 flagMapping);
// the endpoint returns one document whose data.offloadingConfigurations array
// is flattened into individual items. Runs once (no pipeline).
package list

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/datahub"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get offloading configurations
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get offloading configurations",
		Long:  `Get offloading configurations`,
		Example: heredoc.Doc(`
$ c8y datahub configuration list
List the datahub offloading configurations
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("lastMaxReportedUUID", "", "Last max reported UUID")
	cmd.Flags().String("locale", "en", "Locale")

	flags.WithOptions(
		cmd,
		flags.WithCollectionProperty("data.offloadingConfigurations"),
		flags.WithPowershellName("Get-DataHubConfigurationCollection"),
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

	client, err := r.Client()
	if err != nil {
		return err
	}

	common, err := r.Config.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := datahub.ListOffloadingConfigurationsOptions{
			LastMaxReportedUUID: in.String("lastMaxReportedUUID"),
			Locale:              in.String("locale"),
			Limit:               int(common.PageSize),
		}
		return func(ctx context.Context) any {
			return output.FromIterator(client.DataHub.ListOffloadingConfigurations(ctx, opt).Items())
		}, nil
	})
}
