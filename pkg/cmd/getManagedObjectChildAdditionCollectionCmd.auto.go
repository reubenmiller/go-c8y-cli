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

type getManagedObjectChildAdditionCollectionCmd struct {
	*baseCmd
}

func newGetManagedObjectChildAdditionCollectionCmd() *getManagedObjectChildAdditionCollectionCmd {
	ccmd := &getManagedObjectChildAdditionCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "listChildAdditions",
		Short: "Get a collection of managedObjects child additions",
		Long:  `Get a collection of managedObjects child additions`,
		Example: `
$ c8y inventoryReferences listChildAdditions --id 12345
Get a list of the child additions of an existing managed object
        `,
		PreRunE: nil,
		RunE:    ccmd.getManagedObjectChildAdditionCollection,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Managed object id. (required)")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getManagedObjectChildAdditionCollectionCmd) getManagedObjectChildAdditionCollection(cmd *cobra.Command, args []string) error {

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

	path := replacePathParameters("inventory/managedObjects/{id}/childAdditions", pathParameters)

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
