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

type getNewDeviceRequestCmd struct {
	*baseCmd
}

func newGetNewDeviceRequestCmd() *getNewDeviceRequestCmd {
	ccmd := &getNewDeviceRequestCmd{}

	cmd := &cobra.Command{
		Use:   "getNewDeviceRequest",
		Short: "Get a new device requests",
		Long:  `Get a new device requests`,
		Example: `
$ c8y devices getNewDeviceRequest
Get a new device request
        `,
		PreRunE: nil,
		RunE:    ccmd.getNewDeviceRequest,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "New Device Request ID (required)")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getNewDeviceRequestCmd) getNewDeviceRequest(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	commonOptions.AddQueryParameters(&query)
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
		Method:       "GET",
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
