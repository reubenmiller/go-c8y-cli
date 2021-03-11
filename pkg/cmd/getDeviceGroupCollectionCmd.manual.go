package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type getDeviceGroupCollectionCmd struct {
	*baseCmd
}

func NewGetDeviceGroupCollectionCmd() *getDeviceGroupCollectionCmd {
	ccmd := &getDeviceGroupCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "listDeviceGroups",
		Short: "Get device group collection",
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
	inputIterators := &flags.RequestInputIterators{}

	// query parameters
	query := flags.NewQueryTemplate()

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	commonOptions.ResultProperty = "managedObjects"
	commonOptions.AddQueryParameters(query)

	c8yQueryParts, err := flags.WithC8YQueryOptions(
		cmd,
		flags.WithC8YQueryFixedString("(has(c8y_IsDeviceGroup))"),
		flags.WithC8YQueryFormat("name", "(name eq '%s')"),
		flags.WithC8YQueryFormat("type", "(type eq '%s')"),
		flags.WithC8YQueryFormat("fragmentType", "has(%s)"),
		flags.WithC8YQueryFormat("owner", "(owner eq '%s')"),
		flags.WithC8YQueryBool("excludeRootGroup", "not(type eq 'c8y_DeviceGroup')"),
		flags.WithC8YQueryFormat("query", "%s"),
	)

	if err != nil {
		return err
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

	query.SetVariable("query", fmt.Sprintf("$filter=%s+$orderby=%s", filter, orderBy))

	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithBoolValue("withParents", "withParents"),
	)

	if err != nil {
		return nil
	}

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewInitializedMapBuilder()

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects")

	req := c8y.RequestOptions{
		Method:       "GET",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: globalFlagIgnoreAccept,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
