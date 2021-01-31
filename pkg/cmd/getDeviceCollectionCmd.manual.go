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

type getDeviceCollectionCmd struct {
	*baseCmd
}

func NewGetDeviceCollectionCmd() *getDeviceCollectionCmd {
	ccmd := &getDeviceCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get a collection of devices based on filter parameters",
		Long:  `Get a collection of devices based on filter parameters`,
		Example: `
		c8y devices list --name "sensor*" --type myType

		Get a collection of devices of type "myType", and their names start with "sensor"
		`,
		RunE: ccmd.getDeviceCollection,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Device name.")
	cmd.Flags().String("type", "", "Device type.")
	cmd.Flags().Bool("agents", false, "Only include agents.")
	cmd.Flags().String("fragmentType", "", "Device fragment type.")
	cmd.Flags().String("owner", "", "Device owner.")
	cmd.Flags().String("query", "", "Additional query filter")
	cmd.Flags().String("orderBy", "name", "Order by. e.g. _id asc or name asc or creationTime.date desc")
	cmd.Flags().Bool("withParents", false, "include a flat list of all parents and grandparents of the given object")

	// Required flags
	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getDeviceCollectionCmd) getDeviceCollection(cmd *cobra.Command, args []string) error {

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

	c8yQueryParts = append(c8yQueryParts, "(has(c8y_IsDevice) or has(c8y_ModbusDevice))")

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

	if v, err := cmd.Flags().GetBool("agents"); err == nil {
		if v {
			c8yQueryParts = append(c8yQueryParts, "has(com_cumulocity_model_Agent)")
		}
	}

	if v, err := cmd.Flags().GetString("query"); err == nil {
		if v != "" {
			c8yQueryParts = append(c8yQueryParts, v)
		}
	}

	// Compile query
	// replace all spaces with "+" due to url encoding
	filter := url.QueryEscape(strings.Join(c8yQueryParts, " and "))
	orderBy := "name"

	if v, err := cmd.Flags().GetString("orderBy"); err == nil {
		if v != "" {
			orderBy = url.QueryEscape(v)
		}
	}

	// q will automatically add a fragmentType=c8y_IsDevice to the query
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
