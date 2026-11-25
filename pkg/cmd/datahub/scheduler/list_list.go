// v2-based datahub scheduler list: lists the DataHub scheduler's latest job
// runs via DataHub.ListLatestJobs. The report is a single document
// (collectionProperty "-": rendered as-is, no flattening); the management API
// pages with "limit", so the CLI's pageSize maps onto it (the v1 flagMapping).
// Runs once (no pipeline).

package scheduler

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/datahub"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to List scheduler items
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List scheduler items",
		Long:  `List scheduler items`,
		Example: heredoc.Doc(`
$ c8y datahub scheduler list
List the datahub scheduled items
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("jobType", "COMPACTION", "Job type")
	cmd.Flags().Int("offset", 0, "Offset")
	cmd.Flags().Int("nextOffset", 0, "Next offset")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("jobType", "COMPACTION"),
	)

	flags.WithOptions(
		cmd,
		flags.WithCollectionProperty("-"),
		flags.WithPowershellName("Get-DataHubSchedulerCollection"),
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

	// v1 emitted offset/nextOffset only when set (explicit 0 included); the
	// omitempty options struct drops a zero, so read the changed state once.
	offset, _ := cmd.Flags().GetInt("offset")
	nextOffset, _ := cmd.Flags().GetInt("nextOffset")

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := datahub.ListLatestJobsOptions{
			JobType:    in.String("jobType"),
			Offset:     offset,
			NextOffset: nextOffset,
			Limit:      common.PageSize,
		}
		return func(ctx context.Context) any {
			return client.DataHub.ListLatestJobs(ctx, opt)
		}, nil
	})
}
