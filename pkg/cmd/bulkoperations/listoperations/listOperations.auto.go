// Code generated from specification version 1.0.0: DO NOT EDIT
package listoperations

import (
	"fmt"
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
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

	cmd.Flags().String("id", "", "Bulk operation id. (required) (accepts pipeline)")
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
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListOperationsCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
		flags.WithStringValue("id", "bulkOperationId"),
		flags.WithEncodedRelativeTimestamp("dateFrom", "dateFrom", ""),
		flags.WithEncodedRelativeTimestamp("dateTo", "dateTo", ""),
		flags.WithStringValue("status", "status"),
		flags.WithBoolValue("revert", "revert", ""),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}
	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParameters(query)

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("devicecontrol/operations")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "GET",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		DryRun:       cfg.DryRun(),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
