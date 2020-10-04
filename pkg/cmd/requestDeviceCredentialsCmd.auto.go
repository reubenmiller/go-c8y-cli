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

type requestDeviceCredentialsCmd struct {
	*baseCmd
}

func newRequestDeviceCredentialsCmd() *requestDeviceCredentialsCmd {
	ccmd := &requestDeviceCredentialsCmd{}

	cmd := &cobra.Command{
		Use:   "requestDeviceCredentials",
		Short: "Request credentials for a new device",
		Long:  `Device credentials can be enquired by devices that do not have credentials for accessing a tenant yet. Since the device does not have credentials yet, a set of fixed credentials is used for this API. The credentials can be obtained by contacting support. Do not use your tenant credentials with this API.`,
		Example: `
$ c8y devices requestDeviceCredentials --id "device-AD76-matrixer"
Request credentials for a new device
		`,
		RunE: ccmd.requestDeviceCredentials,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Device identifier. Max: 1000 characters. E.g. IMEI (required)")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *requestDeviceCredentialsCmd) requestDeviceCredentials(cmd *cobra.Command, args []string) error {

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
	if v, err := cmd.Flags().GetString("id"); err == nil {
		if v != "" {
			body.Set("id", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "id", err))
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("devicecontrol/deviceCredentials", pathParameters)

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
