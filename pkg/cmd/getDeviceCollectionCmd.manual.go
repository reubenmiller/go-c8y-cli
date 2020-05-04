package cmd

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/fatih/color"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

type getDeviceCollectionCmd struct {
	*baseCmd
}

func newGetDeviceCollectionCmd() *getDeviceCollectionCmd {
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
	cmd.Flags().Bool("withParents", false, "include a flat list of all parents and grandparents of the given object")

	// Required flags
	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getDeviceCollectionCmd) getDeviceCollection(cmd *cobra.Command, args []string) error {

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

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
	orderBy := url.QueryEscape("name")
	query.Add("query", fmt.Sprintf("$filter=%s+$orderby=%s", filter, orderBy))

	if v, err := cmd.Flags().GetBool("withParents"); err == nil {
		if v {
			query.Add("withParents", "true")
		}
	}

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

	queryValue, err := url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// body
	var body map[string]interface{}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("inventory/managedObjects", pathParameters)

	return n.doGetDeviceCollection("GET", path, queryValue, body)
}

func (n *getDeviceCollectionCmd) doGetDeviceCollection(method string, path string, query string, body map[string]interface{}) error {
	resp, err := client.SendRequest(
		context.Background(),
		c8y.RequestOptions{
			Method:       method,
			Path:         path,
			Query:        query,
			Body:         body,
			IgnoreAccept: false,
			DryRun:       globalFlagDryRun,
		})

	if err != nil {
		color.Set(color.FgRed, color.Bold)
	}

	if resp != nil && resp.JSONData != nil {
		if globalFlagPrettyPrint {
			fmt.Printf("%s\n", pretty.Pretty([]byte(*resp.JSONData)))
		} else {
			fmt.Printf("%s\n", *resp.JSONData)
		}
	}

	color.Unset()

	if err != nil {
		return newSystemError("command failed", err)
	}
	return nil
}
