// v2-based operations list: fills the typed operations.ListOptions; the device
// and agent references are resolved by go-c8y. Time-keyset pagination.
package list

import (
	"strconv"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/operations"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/types"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get operation collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get operation collection",
		Long:  `Get a collection of operations based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y operations list --device 12345 --status PENDING
Get pending operations for a device

$ c8y devices list --type myType | c8y operations list --status EXECUTING
List executing operations for each piped device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device id (accepts pipeline)")
	cmd.Flags().String("agent", "", "Agent id")
	cmd.Flags().String("fragmentType", "", "The type of fragment that must be part of the operation, e.g. c8y_Restart")
	cmd.Flags().String("status", "", "Operation status, one of SUCCESSFUL, FAILED, EXECUTING or PENDING")
	cmd.Flags().String("bulkOperationId", "", "Only retrieve operations related to the given bulk operation id")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of operation")
	cmd.Flags().String("dateTo", "", "End date or date and time of operation")
	cmd.Flags().Bool("revert", false, "Sort operations newest to oldest. Must be used with dateFrom and/or dateTo parameters")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "PENDING", "EXECUTING", "SUCCESSFUL", "FAILED"),
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDevice("agent", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "deviceId", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("operations"),
		flags.WithPowershellName("Get-OperationCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.operationCollection+json", "application/vnd.com.nsn.cumulocity.operation+json"),
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

	if err := r.InputFlag("device"); err != nil {
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
		opt := operations.ListOptions{
			FragmentType: in.String("fragmentType"),
			Status:       types.OperationStatus(in.String("status")),
			DateFrom:     in.TimeValue("dateFrom"),
			DateTo:       in.TimeValue("dateTo"),
			Revert:       in.Bool("revert"),
		}
		if device := in.String("device"); device != "" {
			opt.DeviceID = managedobjects.DeviceRef(c8ystream.NameOrID(device))
		}
		if agent := in.String("agent"); agent != "" {
			opt.AgentID = managedobjects.DeviceRef(c8ystream.NameOrID(agent))
		}
		if v := in.String("bulkOperationId"); v != "" {
			if id, convErr := strconv.Atoi(v); convErr == nil {
				opt.BulkOperationID = id
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
		return c8ystream.ListCall(rawOutput, opt, client.Operations.ListAll), nil
	})
}
