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

type getCurrentApplicationUserCollectionCmd struct {
	*baseCmd
}

func newGetCurrentApplicationUserCollectionCmd() *getCurrentApplicationUserCollectionCmd {
	ccmd := &getCurrentApplicationUserCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "listSubscriptions",
		Short: "Get current application subscriptions",
		Long:  `Required authentication with bootstrap user`,
		Example: `
$ c8y currentApplication listSubscriptions
List the current application users/subscriptions
        `,
		PreRunE: nil,
		RunE:    ccmd.getCurrentApplicationUserCollection,
	}

	cmd.SilenceUsage = true

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getCurrentApplicationUserCollectionCmd) getCurrentApplicationUserCollection(cmd *cobra.Command, args []string) error {

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

	path := replacePathParameters("/application/currentApplication/subscriptions", pathParameters)

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
