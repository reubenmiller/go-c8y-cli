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

type queryManagedObjectCollectionCmd struct {
	*baseCmd
}

func newQueryManagedObjectCollectionCmd() *queryManagedObjectCollectionCmd {
	ccmd := &queryManagedObjectCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "find",
		Short: "Get a collection of managedObjects based on Cumulocity query language",
		Long:  `Get a collection of managedObjects based on Cumulocity query language`,
		Example: `
$ c8y inventory find --query "name eq 'roomUpperFloor_*'"
Get a list of managed objects
		`,
		RunE: ccmd.queryManagedObjectCollection,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("query", "", "ManagedObject query. (required)")
	cmd.Flags().String("orderBy", "", "Order the results by the given parameter. i.e. 'id asc'")
	cmd.Flags().Bool("withParents", false, "include a flat list of all parents and grandparents of the given object")

	// Required flags
	cmd.MarkFlagRequired("query")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *queryManagedObjectCollectionCmd) queryManagedObjectCollection(cmd *cobra.Command, args []string) error {

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	commonOptions.ResultProperty = "managedObjects"
	commonOptions.AddQueryParameters(&query)

	orderBy := ""
	if v, err := cmd.Flags().GetString("orderBy"); err == nil {
		if v != "" {
			orderBy = v
		}
	}
	if v, err := cmd.Flags().GetString("query"); err == nil {
		if v != "" {
			c8yQuery := fmt.Sprintf("$filter=%s", url.QueryEscape(v))

			if orderBy != "" {
				c8yQuery = c8yQuery + fmt.Sprintf("+$orderby=%s", url.QueryEscape(orderBy))
			}

			query.Add("query", c8yQuery)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "query", err))
	}
	if cmd.Flags().Changed("withParents") {
		if v, err := cmd.Flags().GetBool("withParents"); err == nil {
			query.Add("withParents", fmt.Sprintf("%v", v))
		} else {
			return newUserError("Flag does not exist")
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

	path := replacePathParameters("inventory/managedObjects", pathParameters)

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
