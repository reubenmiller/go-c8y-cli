// v2-based bulk operations list: fills the typed bulkoperations.ListOptions
// (withDeleted / dateFrom / dateTo / generalStatus filters). No pipeline input,
// so the command runs once. Time-keyset pagination via the shared strategy.
package list

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/bulkoperations"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get bulk operation collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get bulk operation collection",
		Long:  `Get a collection of bulk operations`,
		Example: heredoc.Doc(`
$ c8y bulkoperations list
Get a list of bulk operations

$ c8y bulkoperations list --dateFrom -1d
Get a list of bulk operations created in the last 1 day

$ c8y bulkoperations list --status SCHEDULED --status EXECUTING
Get a list of bulk operations in the general status SCHEDULED or EXECUTING
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().Bool("withDeleted", false, "Include CANCELLED bulk operations")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of the bulk operation")
	cmd.Flags().String("dateTo", "", "End date or date and time of the bulk operation")
	cmd.Flags().StringSlice("status", []string{""}, "Operation status, can be one of SUCCESSFUL, FAILED, EXECUTING or PENDING.")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "CANCELED", "SCHEDULED", "EXECUTING", "EXECUTING_WITH_ERROR", "FAILED"),
	)

	flags.WithOptions(
		cmd,
		flags.WithCollectionProperty("bulkOperations"),
		flags.WithPowershellName("Get-BulkOperationCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.bulkOperationCollection+json", "application/vnd.com.nsn.cumulocity.bulkoperation+json"),
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

	rawOutput := r.Config.RawOutput()
	paginationStrategy := pagination.StrategyKind(r.Config.PaginationStrategy())

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := bulkoperations.ListOptions{
			WithDeleted: in.Bool("withDeleted"),
			DateFrom:    in.TimeValue("dateFrom"),
			DateTo:      in.TimeValue("dateTo"),
		}
		for _, status := range in.StringSlice("status") {
			if status != "" {
				opt.GeneralStatus = append(opt.GeneralStatus, status)
			}
		}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.BulkOperations.ListAll), nil
	})
}
