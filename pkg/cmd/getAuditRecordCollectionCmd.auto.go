// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// GetAuditRecordCollectionCmd command
type GetAuditRecordCollectionCmd struct {
	*baseCmd
}

// NewGetAuditRecordCollectionCmd creates a command to Get audit record collection
func NewGetAuditRecordCollectionCmd() *GetAuditRecordCollectionCmd {
	ccmd := &GetAuditRecordCollectionCmd{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get audit record collection",
		Long: `Audit records contain information about modifications to other Cumulocity entities. For example the audit records contain each operation state transition, so they can be used to check when an operation transitioned from PENDING -> EXECUTING -> SUCCESSFUL.
`,
		Example: `
$ c8y auditRecords list --pageSize 100
Get a list of audit records
        `,
		PreRunE: nil,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("source", "", "Source Id or object containing an .id property of the element that should be detected. i.e. AlarmID, or Operation ID. Note: Only one source can be provided (accepts pipeline)")
	cmd.Flags().String("type", "", "Type")
	cmd.Flags().String("user", "", "Username")
	cmd.Flags().String("application", "", "Application")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of audit record occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of audit record occurrence.")
	cmd.Flags().Bool("revert", false, "Return the newest instead of the oldest audit records. Must be used with dateFrom and dateTo parameters")

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("source", "source", false, "id", "source.id", "managedObject.id"),
		flags.WithCollectionProperty("auditRecords"),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetAuditRecordCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
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
		flags.WithStringValue("source", "source"),
		flags.WithStringValue("type", "type"),
		flags.WithStringValue("user", "user"),
		flags.WithStringValue("application", "application"),
		flags.WithRelativeTimestamp("dateFrom", "dateFrom", ""),
		flags.WithRelativeTimestamp("dateTo", "dateTo", ""),
		flags.WithBoolValue("revert", "revert", ""),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}
	commonOptions, err := getCommonOptions(cmd)
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
	path := flags.NewStringTemplate("/audit/auditRecords")
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
		IgnoreAccept: globalFlagIgnoreAccept,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
