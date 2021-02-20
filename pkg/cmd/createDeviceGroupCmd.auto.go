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

type CreateDeviceGroupCmd struct {
	*baseCmd
}

func NewCreateDeviceGroupCmd() *CreateDeviceGroupCmd {
	ccmd := &CreateDeviceGroupCmd{}
	cmd := &cobra.Command{
		Use:   "createGroup",
		Short: "Create device group",
		Long: `Create a new device group to logically group one or more devices
`,
		Example: `
$ c8y devices createGroup --name mygroup
Create device group

$ c8y devices createGroup --name mygroup --data "custom_value1=1234"
Create device group with custom properties
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Device group name (accepts pipeline)")
	cmd.Flags().String("type", "", "Device group type (c8y_DeviceGroup (root folder) or c8y_DeviceSubGroup (sub folder)). Defaults to c8y_DeviceGroup")
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("type", "c8y_DeviceGroup", "c8y_DeviceSubGroup"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("name", "name", false),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *CreateDeviceGroupCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("type", "type"),
		flags.WithDefaultTemplateString(`
{type: 'c8y_DeviceGroup', c8y_IsDeviceGroup: {}}`),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
		flags.WithRequiredProperties("name"),
	)
	if err != nil {
		return newUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects")
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
