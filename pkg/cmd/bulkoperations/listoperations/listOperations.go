// v2-based bulk operation operations list: the id flag (bulkOperationId) drives
// iteration (pipe a bulk operation / --id slice). Each id fills
// BulkOperations.ListOperations, which reads the operations collection
// (devicecontrol/operations) scoped by the bulkOperationId query parameter.
package listoperations

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

// ListOperationsCmd command
type ListOperationsCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListOperationsCmd creates a command to Get operations collection
func NewListOperationsCmd(f *cmdutil.Factory) *ListOperationsCmd {
	ccmd := &ListOperationsCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "listOperations",
		Short: "Get operations collection",
		Long:  `Get a collection of operations related to a bulk operation`,
		Example: heredoc.Doc(`
$ c8y bulkoperations listOperations --id 10 --status PENDING
Get a list of pending operations from bulk operation with id 10

$ c8y bulkoperations list --filter "status eq IN_PROGRESS" | c8y bulkoperations listOperations --status PENDING
Get all pending operations from all bulk operations which are still in progress (using pipeline)

$ c8y bulkoperations list | c8y bulkoperations listOperations --status EXECUTING --dateTo "-10d" | c8y operations update --status FAILED --failureReason "Manually cancelled stale operation"
Check all bulk operations if they have any related operations still in executing state and were created more than 10 days ago, then cancel it with a custom message
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Bulk operation id. (required) (accepts pipeline)")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of operation.")
	cmd.Flags().String("dateTo", "", "End date or date and time of operation.")
	cmd.Flags().String("status", "", "Operation status, can be one of SUCCESSFUL, FAILED, EXECUTING or PENDING.")
	cmd.Flags().Bool("revert", false, "Sort operations newest to oldest. Must be used with dateFrom and/or dateTo parameters")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "PENDING", "EXECUTING", "SUCCESSFUL", "FAILED"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "bulkOperationId", true, "id", "bulkOperationId"),
		flags.WithCollectionProperty("operations"),
		flags.WithPowershellName("Get-BulkOperationOperationCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.operationCollection+json", "application/vnd.com.nsn.cumulocity.operation+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListOperationsCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
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
		opt := bulkoperations.ListOperationsOptions{
			BulkOperationID: in.String("id"),
			DateFrom:        in.TimeValue("dateFrom"),
			DateTo:          in.TimeValue("dateTo"),
			Status:          in.String("status"),
			Revert:          in.Bool("revert"),
		}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.BulkOperations.ListAllOperations), nil
	})
}
