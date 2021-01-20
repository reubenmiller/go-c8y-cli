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

type deleteApplicationBinaryCmd struct {
	*baseCmd
}

func newDeleteApplicationBinaryCmd() *deleteApplicationBinaryCmd {
	ccmd := &deleteApplicationBinaryCmd{}

	cmd := &cobra.Command{
		Use:   "deleteApplicationBinary",
		Short: "Remove application binary",
		Long: `Remove an application binaries related to the given application
The active version can not be deleted and the server will throw an error if you try.
`,
		Example: `
$ c8y applications deleteApplicationBinary --application 12345 --binaryId 9876
Remove an application binary related to a Hosted (web) application
        `,
		PreRunE: validateDeleteMode,
		RunE:    ccmd.deleteApplicationBinary,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("application", "", "Application id (required)")
	cmd.Flags().StringSlice("binaryId", []string{""}, "Application binary id (required)")
	addProcessingModeFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("application")
	cmd.MarkFlagRequired("binaryId")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *deleteApplicationBinaryCmd) deleteApplicationBinary(cmd *cobra.Command, args []string) error {

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
	if cmd.Flags().Changed("processingMode") {
		if v, err := cmd.Flags().GetString("processingMode"); err == nil && v != "" {
			headers.Add("X-Cumulocity-Processing-Mode", v)
		}
	}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewMapBuilder()

	// path parameters
	pathParameters := make(map[string]string)
	if cmd.Flags().Changed("application") {
		applicationInputValues, applicationValue, err := getApplicationSlice(cmd, args, "application")

		if err != nil {
			return newUserError("no matching applications found", applicationInputValues, err)
		}

		if len(applicationValue) == 0 {
			return newUserError("no matching applications found", applicationInputValues)
		}

		for _, item := range applicationValue {
			if item != "" {
				pathParameters["application"] = newIDValue(item).GetID()
			}
		}
	}
	if items, err := cmd.Flags().GetStringSlice("binaryId"); err == nil {
		if len(items) > 0 {
			for _, v := range items {
				if v != "" {
					pathParameters["binaryId"] = v
				}
			}
		}
	} else {
		return newUserError("Flag does not exist")
	}

	path := replacePathParameters("/application/applications/{application}/binaries/{binaryId}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "DELETE",
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
