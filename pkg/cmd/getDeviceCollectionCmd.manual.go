package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type GetDeviceCollectionCmd struct {
	*baseCmd
}

func NewGetDeviceCollectionCmd() *GetDeviceCollectionCmd {
	ccmd := &GetDeviceCollectionCmd{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get device collection",
		Long:  `Get a collection of devices based on filter parameters`,
		Example: heredoc.Doc(`
		c8y devices list --name "sensor*" --type myType

		Get a collection of devices of type "myType", and their names start with "sensor"
		`),
		RunE: ccmd.RunE,
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

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport(""),
	)

	// Required flags
	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *GetDeviceCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithC8YQueryFormat("name", "(name eq '%s')"),
		flags.WithC8YQueryFormat("type", "(type eq '%s')"),
		flags.WithC8YQueryFormat("fragmentType", "has(%s)"),
		flags.WithC8YQueryFormat("owner", "(owner eq '%s')"),
		flags.WithC8YQueryBool("agents", "has(com_cumulocity_model_Agent)"),
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

	// q will automatically add a fragmentType=c8y_IsDevice to the query
	query.SetVariable("q", fmt.Sprintf("$filter=%s+$orderby=%s", filter, orderBy))

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
		DryRun:       cliConfig.DryRun(),
		IgnoreAccept: cliConfig.IgnoreAcceptHeader(),
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
