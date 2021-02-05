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

type CreateDeviceCmd struct {
	*baseCmd
}

func NewCreateDeviceCmd() *CreateDeviceCmd {
	ccmd := &CreateDeviceCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a device",
		Long: `Create a device (managed object) with the special c8y_IsDevice fragment.
`,
		Example: `
$ c8y devices create --name myDevice
Create device

$ c8y devices create --name myDevice --data "custom_value1=1234"
Create device with custom properties
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Device name")
	cmd.Flags().String("type", "", "Device type")
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport(""),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *CreateDeviceCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	err = flags.WithQueryParameters(
		cmd,
		query,
	)
	if err != nil {
		return newUserError(err)
	}
	err = flags.WithQueryOptions(
		cmd,
		query,
	)
	if err != nil {
		return newUserError(err)
	}

	queryValue, err = url.QueryUnescape(query.Encode())

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
		flags.WithDataValue(FlagDataName, ""),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("type", "type"),
	)
	if err != nil {
		return newUserError(err)
	}

	bodyErr := body.MergeJsonnet(`
{  c8y_IsDevice: {},
}
`, false)
	if bodyErr != nil {
		return newSystemError("Template error. ", bodyErr)
	}
	if err := setLazyDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
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

	path := replacePathParameters("inventory/managedObjects", pathParameters)

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

	return processRequestAndResponseWithWorkers(cmd, &req, PipeOption{"", false})
}
