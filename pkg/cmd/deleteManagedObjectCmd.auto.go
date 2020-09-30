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

type deleteManagedObjectCmd struct {
	*baseCmd
}

func newDeleteManagedObjectCmd() *deleteManagedObjectCmd {
	ccmd := &deleteManagedObjectCmd{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete inventory/s",
		Long:  `Delete a managed object by id`,
		Example: `
$ c8y inventory delete --id 12345
Delete a managed object

$ c8y inventory delete --id 12345 --cascade
Delete a managed object
		`,
		RunE: ccmd.deleteManagedObject,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "ManagedObject id (required)")
	cmd.Flags().Bool("cascade", false, "Remove all child devices and child assets will be deleted recursively. By default, the delete operation is propagated to the subgroups only if the deleted object is a group")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *deleteManagedObjectCmd) deleteManagedObject(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
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

	// path parameters
	pathParameters := make(map[string]string)
	if v, err := cmd.Flags().GetString("id"); err == nil {
		if v != "" {
			pathParameters["id"] = v
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "id", err))
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
