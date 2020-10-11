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

type createDeviceCmd struct {
	*baseCmd
}

func newCreateDeviceCmd() *createDeviceCmd {
	ccmd := &createDeviceCmd{}

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
		RunE: ccmd.createDevice,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("name", []string{""}, "Device name (required)")
	cmd.Flags().String("type", "", "Device type")
	addDataFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("name")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *createDeviceCmd) createDevice(cmd *cobra.Command, args []string) error {

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
	if items, err := cmd.Flags().GetStringSlice("name"); err == nil {
		if len(items) > 0 {
			for _, v := range items {
				if v != "" {
					body.Set("name", v)
				}
			}
		}
	} else {
		return newUserError("Flag does not exist")
	}
	if v, err := cmd.Flags().GetString("type"); err == nil {
		if v != "" {
			body.Set("type", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "type", err))
	}
	bodyErr := body.MergeJsonnet(`
{  c8y_IsDevice: {},
}
`, false)
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
