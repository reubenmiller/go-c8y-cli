// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type NewAuditCmd struct {
	*baseCmd
}

func NewNewAuditCmd() *NewAuditCmd {
	ccmd := &NewAuditCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new audit record",
		Long:  `Create a new audit record for a given action`,
		Example: `
$ c8y auditRecords create --type "ManagedObject" --time "0s" --text "Managed Object updated: my_Prop: value" --source $Device.id --activity "Managed Object updated" --severity "information"
Create an audit record for a custom managed object update
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("type", "", "Identifies the type of this audit record. (required)")
	cmd.Flags().String("time", "0s", "Time of the audit record. Defaults to current timestamp.")
	cmd.Flags().String("text", "", "Text description of the audit record. (required)")
	cmd.Flags().String("source", "", "An optional ManagedObject that the audit record originated from (required)")
	cmd.Flags().String("activity", "", "The activity that was carried out. (required)")
	cmd.Flags().String("severity", "", "The severity of action: critical, major, minor, warning or information. (required)")
	cmd.Flags().String("user", "", "The user responsible for the audited action.")
	cmd.Flags().String("application", "", "The application used to carry out the audited action.")
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("", "", false),
	)

	// Required flags
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("text")
	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("activity")
	cmd.MarkFlagRequired("severity")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *NewAuditCmd) RunE(cmd *cobra.Command, args []string) error {
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
	)
	if err != nil {
		return newUserError(err)
	}

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		flags.WithProcessingModeValue(),
	)
	if err != nil {
		return newUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
	)
	if err != nil {
		return newUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		WithDataValue(),
		flags.WithStringValue("type", "type"),
		flags.WithRelativeTimestamp("time", "time", ""),
		flags.WithStringValue("text", "text"),
		flags.WithStringValue("source", "source.id"),
		flags.WithStringValue("activity", "activity"),
		flags.WithStringValue("severity", "severity"),
		flags.WithStringValue("user", "user"),
		flags.WithStringValue("application", "application"),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
	)
	if err != nil {
		return newUserError(err)
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
		Method:       "POST",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
