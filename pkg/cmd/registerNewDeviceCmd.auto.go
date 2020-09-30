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

type registerNewDeviceCmd struct {
	*baseCmd
}

func newRegisterNewDeviceCmd() *registerNewDeviceCmd {
	ccmd := &registerNewDeviceCmd{}

	cmd := &cobra.Command{
		Use:   "registerNewDevice",
		Short: "Register a new device (request)",
		Long:  `Register a new device (request)`,
		Example: `
$ c8y devices registerNewDevice --id "ASDF098SD1J10912UD92JDLCNCU8"
Register a new device request
		`,
		RunE: ccmd.registerNewDevice,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Device identifier. Max: 1000 characters. E.g. IMEI (required)")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *registerNewDeviceCmd) registerNewDevice(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	if cmd.Flags().Changed("pageSize") {
		if v, err := cmd.Flags().GetInt("pageSize"); err == nil && v > 0 {
			query.Add("pageSize", fmt.Sprintf("%d", v))
		}
	}

	if cmd.Flags().Changed("withTotalPages") {
		if v, err := cmd.Flags().GetBool("withTotalPages"); err == nil && v {
			query.Add("withTotalPages", "true")
		}
	}
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

	path := replacePathParameters("devicecontrol/newDeviceRequests", pathParameters)

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
