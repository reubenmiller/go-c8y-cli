// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"

	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type ApproveNewDeviceRequestCmd struct {
	*baseCmd
}

func NewApproveNewDeviceRequestCmd() *ApproveNewDeviceRequestCmd {
	ccmd := &ApproveNewDeviceRequestCmd{}
	cmd := &cobra.Command{
		Use:   "approveDeviceRequest",
		Short: "Approve a new device request",
		Long:  `Approve a new device request. Note: a device can only be approved if the platform has received a request for device credentials.`,
		Example: `
$ c8y devices approveDeviceRequest --id "1234010101s01ldk208"
Approve a new device request
        `,
		PreRunE: validateUpdateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Device identifier (required) (accepts pipeline)")
	cmd.Flags().String("status", "ACCEPTED", "Status of registration")
	addProcessingModeFlag(cmd)

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "ACCEPTED"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *ApproveNewDeviceRequestCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithStringValue("status", "status"),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
	)
	if err != nil {
		return newUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("devicecontrol/newDeviceRequests/{id}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		flags.WithStringValue("id", "id"),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "PUT",
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
