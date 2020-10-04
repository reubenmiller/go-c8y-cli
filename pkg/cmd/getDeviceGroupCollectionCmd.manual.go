package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type getDeviceGroupCollectionCmd struct {
	*baseCmd
}

func newGetDeviceGroupCollectionCmd() *getDeviceGroupCollectionCmd {
	ccmd := &getDeviceGroupCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "listDeviceGroups",
		Short: "Get a collection of device groups based on filter parameters",
		Long:  `Get a collection of device groups based on filter parameters`,
		Example: `
		c8y devices listDeviceGroups --name "MyGroup*"

		Get a collection of device groups with names that start with "MyGroup"
		`,
		RunE: ccmd.getDeviceGroupCollection,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Device group name.")
	cmd.Flags().String("type", "", "Device group type.")
	cmd.Flags().String("fragmentType", "", "Device group fragment type.")
	cmd.Flags().String("owner", "", "Device group owner.")
	cmd.Flags().String("query", "", "Additional query filter")
	cmd.Flags().Bool("excludeRootGroup", false, "Exclude root groups from the list")
	cmd.Flags().Bool("withParents", false, "include a flat list of all parents and grandparents of the given object")

	// Required flags
	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getDeviceGroupCollectionCmd) getDeviceGroupCollection(cmd *cobra.Command, args []string) error {

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	commonOptions.ResultProperty = "managedObjects"
	commonOptions.AddQueryParameters(&query)

	var c8yQueryParts = make([]string, 0)

	c8yQueryParts = append(c8yQueryParts, "(has(c8y_IsDeviceGroup))")

	if v, err := cmd.Flags().GetString("name"); err == nil {
		if v != "" {
			c8yQueryParts = append(c8yQueryParts, fmt.Sprintf("(name eq '%s')", v))
		}
	}

	if v, err := cmd.Flags().GetString("type"); err == nil {
		if v != "" {
			c8yQueryParts = append(c8yQueryParts, fmt.Sprintf("(type eq '%s')", v))
		}
	}

	if v, err := cmd.Flags().GetString("fragmentType"); err == nil {
		if v != "" {
			c8yQueryParts = append(c8yQueryParts, fmt.Sprintf("has(%s)", v))
		}
	}

	if v, err := cmd.Flags().GetString("owner"); err == nil {
		if v != "" {
			c8yQueryParts = append(c8yQueryParts, fmt.Sprintf("(owner eq '%s')", v))
		}
	}

	if v, err := cmd.Flags().GetString("query"); err == nil {
		if v != "" {
			c8yQueryParts = append(c8yQueryParts, v)
		}
	}

	if v, err := cmd.Flags().GetBool("excludeRootGroup"); err == nil {
		if v {
			c8yQueryParts = append(c8yQueryParts, "not(type eq 'c8y_DeviceGroup')")
		}
	}

	// Compile query
	// replace all spaces with "+" due to url encoding
	filter := url.QueryEscape(strings.Join(c8yQueryParts, " and "))
	orderBy := url.QueryEscape("name")
	query.Add("query", fmt.Sprintf("$filter=%s+$orderby=%s", filter, orderBy))

	if v, err := cmd.Flags().GetBool("withParents"); err == nil {
		if v {
			query.Add("withParents", "true")
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
