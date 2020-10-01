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

type approveNewDeviceRequestCmd struct {
	*baseCmd
}

func newApproveNewDeviceRequestCmd() *approveNewDeviceRequestCmd {
	ccmd := &approveNewDeviceRequestCmd{}

	cmd := &cobra.Command{
		Use:   "approveDeviceRequest",
		Short: "Approve a new device request",
		Long:  `Approve a new device request. Note: a device can only be approved if the platform has received a request for device credentials.`,
		Example: `
$ c8y devices approveDeviceRequest --id "1234010101s01ldk208"
Approve a new device request
		`,
		RunE: ccmd.approveNewDeviceRequest,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Device identifier (required)")
	cmd.Flags().String("status", "ACCEPTED", "Status of registration")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *approveNewDeviceRequestCmd) approveNewDeviceRequest(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	if cmd.Flags().Changed("pageSize") || globalUseNonDefaultPageSize {
		if v, err := cmd.Flags().GetInt("pageSize"); err == nil && v > 0 {
			query.Add("pageSize", fmt.Sprintf("%d", v))
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
	if v, err := cmd.Flags().GetString("status"); err == nil {
		if v != "" {
			body.Set("status", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "status", err))
	}

	// path parameters
	pathParameters := make(map[string]string)
	if v, err := cmd.Flags().GetString("id"); err == nil {
		if v != "" {
			pathParameters["id"] = v
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "id", err))
	}

	path := replacePathParameters("devicecontrol/newDeviceRequests/{id}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "PUT",
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
