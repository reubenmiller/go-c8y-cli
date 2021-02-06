// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type NewBulkOperationCmd struct {
	*baseCmd
}

func NewNewBulkOperationCmd() *NewBulkOperationCmd {
	ccmd := &NewBulkOperationCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new bulk operation",
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

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("group"),
	)

	// Required flags
	cmd.MarkFlagRequired("creationRampSec")
	cmd.MarkFlagRequired("operation")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *NewBulkOperationCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	// query parameters
	query := url.Values{}
	err = flags.WithQueryParameters(
		cmd,
		query,
	)
	if err != nil {
		return newUserError(err)
	}

	queryValue, err := url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
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
	)
	if err != nil {
		return newUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
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
		return newUserError(err)
	}

	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	err = flags.WithPathParameters(
		cmd,
		pathParameters,
	)
	if err != nil {
		return err
	}

	path := replacePathParameters("devicecontrol/bulkoperations", pathParameters)

	req := c8y.RequestOptions{
		Method:       "POST",
		Path:         path,
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	pipeOption := PipeOption{
		Name:              "group",
		Property:          "groupId",
		Required:          true,
		ResolveByNameType: "devicegroup",
		IteratorType:      "body",
	}
	return processRequestAndResponseWithWorkers(cmd, &req, pipeOption)
}
