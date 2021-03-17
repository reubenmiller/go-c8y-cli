// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// NewAuditCmd command
type NewAuditCmd struct {
	*subcommand.SubCommand
}

// NewNewAuditCmd creates a command to Create audit record
func NewNewAuditCmd() *NewAuditCmd {
	ccmd := &NewAuditCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create audit record",
		Long:  `Create a new audit record for a given action`,
		Example: heredoc.Doc(`
$ c8y auditRecords create --type "ManagedObject" --time "0s" --text "Managed Object updated: my_Prop: value" --source $Device.id --activity "Managed Object updated" --severity "information"
Create an audit record for a custom managed object update
        `),
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("type", "", "Identifies the type of this audit record. (required)")
	cmd.Flags().String("time", "0s", "Time of the audit record. Defaults to current timestamp.")
	cmd.Flags().String("text", "", "Text description of the audit record. (required)")
	cmd.Flags().String("source", "", "An optional ManagedObject that the audit record originated from (required) (accepts pipeline)")
	cmd.Flags().String("activity", "", "The activity that was carried out. (required)")
	cmd.Flags().String("severity", "", "The severity of action: critical, major, minor, warning or information. (required)")
	cmd.Flags().String("user", "", "The user responsible for the audited action.")
	cmd.Flags().String("application", "", "The application used to carry out the audited action.")
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("severity", "critical", "major", "minor", "warning", "information"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("source", "source.id", true, "id"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("text")
	_ = cmd.MarkFlagRequired("activity")
	_ = cmd.MarkFlagRequired("severity")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
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
		return cmderrors.NewUserError(err)
	}

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
		flags.WithProcessingModeValue(),
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
		Method:       "POST",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: cliConfig.IgnoreAcceptHeader(),
		DryRun:       cliConfig.DryRun(),
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
