package list

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdDevicesList struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdDevicesList(f *cmdutil.Factory) *CmdDevicesList {
	ccmd := &CmdDevicesList{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get device collection",
		Long:  `Get a collection of devices based on filter parameters`,
		Example: heredoc.Doc(`
		$ c8y devices list --name "sensor*" --type myType
		Get a collection of devices of type "myType", and their names start with "sensor"

		$ c8y devices list --query "name eq '*sensor*' and creationTime.date gt '2021-04-02T00:00:00'"
		Get devices which names containing 'sensor' and were created after 2021-04-02

		$ echo -e "c8y_MacOS\nc8y_Linux" | c8y devices list --queryTemplate "type eq '%s'"
		Get devices with type 'c8y_MacOS' then devices with type 'c8y_Linux' (using pipeline)
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("type", "", "Filter by type")
	cmd.Flags().Bool("agents", false, "Only include agents.")
	cmd.Flags().String("fragmentType", "", "Filter by fragment type")
	cmd.Flags().String("owner", "", "Filter by owner")
	cmd.Flags().String("query", "", "Additional query filter (accepts pipeline)")
	cmd.Flags().String("queryTemplate", "", "String template to be used when applying the given query. Use %s to reference the query/pipeline input")
	cmd.Flags().String("orderBy", "name", "Order by. e.g. _id asc or name asc or creationTime.date desc")
	cmd.Flags().Bool("withParents", false, "Include a flat list of all parents and grandparents of the given object")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("query", "query", false, "c8y_DeviceQueryString"),
	)

	// Required flags
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdDevicesList) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}

	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()

	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
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
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
		flags.WithCustomStringValue(
			flags.BuildCumulocityQuery(cmd, c8yQueryParts, orderBy),
			func() string { return "q" },
			"query",
		),
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
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

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
		DryRun:       cfg.DryRun(),
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
