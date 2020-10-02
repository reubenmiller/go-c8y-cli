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

type deleteDeviceGroupCmd struct {
	*baseCmd
}

func newDeleteDeviceGroupCmd() *deleteDeviceGroupCmd {
	ccmd := &deleteDeviceGroupCmd{}

	cmd := &cobra.Command{
		Use:   "deleteGroup",
		Short: "Delete device group",
		Long: `Delete an existing device group, and optional
`,
		Example: `
$ c8y devices deleteGroup --id 12345
Get device group by id
		`,
		RunE: ccmd.deleteDeviceGroup,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Device group ID (required)")
	cmd.Flags().Bool("cascade", false, "Remove all child devices and child assets will be deleted recursively. By default, the delete operation is propagated to the subgroups only if the deleted object is a group")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *deleteDeviceGroupCmd) deleteDeviceGroup(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	if cmd.Flags().Changed("cascade") {
		if v, err := cmd.Flags().GetBool("cascade"); err == nil {
			query.Add("cascade", fmt.Sprintf("%v", v))
		} else {
			return newUserError("Flag does not exist")
		}
	}
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
	if cmd.Flags().Changed("id") {
		idInputValues, idValue, err := getFormattedDeviceGroupSlice(cmd, args, "id")

		if err != nil {
			return newUserError("no matching device groups found", idInputValues, err)
		}

		if len(idValue) == 0 {
			return newUserError("no matching device groups found", idInputValues)
		}

		for _, item := range idValue {
			if item != "" {
				pathParameters["id"] = newIDValue(item).GetID()
			}
		}
	}

	path := replacePathParameters("inventory/managedObjects/{id}", pathParameters)

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
