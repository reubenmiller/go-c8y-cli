// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// NewBulkOperationCmd command
type NewBulkOperationCmd struct {
	*baseCmd
}

// NewNewBulkOperationCmd creates a command to Create bulk operation
func NewNewBulkOperationCmd() *NewBulkOperationCmd {
	ccmd := &NewBulkOperationCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create bulk operation",
		Long:  `Create a new bulk operation`,
		Example: `
$ c8y operations create --device mydevice --data "{c8y_Restart:{}}"
Create operation for a device
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Identifies the target group on which this operation should be performed. (required) (accepts pipeline)")
	cmd.Flags().String("startDate", "300s", "Time when operations should be created. Defaults to 300s")
	cmd.Flags().Float32("creationRampSec", 0, "Delay between every operation creation. (required)")
	cmd.Flags().String("operation", "", "Operation prototype to send to each device in the group (required)")
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("group", "groupId", true, "id"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("creationRampSec")
	_ = cmd.MarkFlagRequired("operation")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

// RunE executes the command
func (n *NewBulkOperationCmd) RunE(cmd *cobra.Command, args []string) error {
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
		WithDeviceGroupByNameFirstMatch(args, "group", "groupId"),
		flags.WithRelativeTimestamp("startDate", "startDate", ""),
		flags.WithFloatValue("creationRampSec", "creationRamp"),
		flags.WithDataValue("operation", "operationPrototype"),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
		flags.WithRequiredProperties("groupId", "startDate", "creationRamp", "operationPrototype"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("devicecontrol/bulkoperations")
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
		IgnoreAccept: globalFlagIgnoreAccept,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
