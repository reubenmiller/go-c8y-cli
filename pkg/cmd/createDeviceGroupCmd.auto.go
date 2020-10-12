// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type createDeviceGroupCmd struct {
	*baseCmd
}

func newCreateDeviceGroupCmd() *createDeviceGroupCmd {
	ccmd := &createDeviceGroupCmd{}

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
		RunE:    ccmd.createDeviceGroup,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Device group name (required)")
	cmd.Flags().String("type", "", "Device group type (c8y_DeviceGroup (root folder) or c8y_DeviceSubGroup (sub folder)). Defaults to c8y_DeviceGroup")
	addDataFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("name")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *createDeviceGroupCmd) createDeviceGroup(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	queryValue, err = url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewMapBuilder()
	body.SetMap(getDataFlag(cmd))
	if v, err := cmd.Flags().GetString("name"); err == nil {
		if v != "" {
			body.Set("name", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "name", err))
	}
	if v, err := cmd.Flags().GetString("type"); err == nil {
		if v != "" {
			body.Set("type", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "type", err))
	}
	bodyErr := body.MergeJsonnet(`
{  type: "c8y_DeviceGroup",
  c8y_IsDeviceGroup: {},
}
`, true)
	if bodyErr != nil {
		return newSystemError("Template error. ", bodyErr)
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("inventory/managedObjects", pathParameters)

	req := c8y.RequestOptions{
		Method:       "POST",
		Path:         path,
		Query:        queryValue,
		Body:         body.GetMap(),
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponse([]c8y.RequestOptions{req}, commonOptions)
}
