// Code generated from specification version 1.0.0: DO NOT EDIT
package list

import (
	"fmt"
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get device collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
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

$ c8y devices list --creationTimeDateFrom -7d
Get devices which where registered longer than 7 days ago

$ c8y devices list --creationTimeDateTo -1d
Get devices which where registered in the last day

$ echo -e "c8y_MacOS\nc8y_Linux" | c8y devices list --queryTemplate "type eq '%s'"
Get devices with type 'c8y_MacOS' then devices with type 'c8y_Linux' (using pipeline)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("query", "", "Additional query filter (accepts pipeline)")
	cmd.Flags().String("queryTemplate", "", "String template to be used when applying the given query. Use %s to reference the query/pipeline input")
	cmd.Flags().String("orderBy", "name", "Order by. e.g. _id asc or name asc or creationTime.date desc")
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("type", "", "Filter by type")
	cmd.Flags().Bool("agents", false, "Only include agents")
	cmd.Flags().String("fragmentType", "", "Filter by fragment type")
	cmd.Flags().String("owner", "", "Filter by owner")
	cmd.Flags().String("availability", "", "Filter by c8y_Availability.status")
	cmd.Flags().String("lastMessageDateTo", "", "Filter c8y_Availability.lastMessage to a specific date")
	cmd.Flags().String("lastMessageDateFrom", "", "Filter c8y_Availability.lastMessage from a specific date")
	cmd.Flags().String("creationTimeDateTo", "", "Filter creationTime.date to a specific date")
	cmd.Flags().String("creationTimeDateFrom", "", "Filter creationTime.date from a specific date")
	cmd.Flags().StringSlice("group", []string{""}, "Filter by group inclusion")
	cmd.Flags().Bool("skipChildrenNames", false, "Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved")
	cmd.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.")
	cmd.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")
	cmd.Flags().Bool("withGroups", false, "When set to true it returns additional information about the groups to which the searched managed object belongs. This results in setting the assetParents property with additional information about the groups.")
	cmd.Flags().Bool("withParents", false, "Include a flat list of all parents and grandparents of the given object")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("availability", "AVAILABLE", "UNAVAILABLE", "MAINTENANCE"),
		completion.WithDeviceGroup("group", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,

		flags.WithExtendedPipelineSupport("query", "query", false, "c8y_DeviceQueryString"),
		flags.WithPipelineAliases("lastMessageDateTo", "time", "creationTime", "lastUpdated"),
		flags.WithPipelineAliases("lastMessageDateFrom", "time", "creationTime", "lastUpdated"),
		flags.WithPipelineAliases("creationTimeDateTo", "time", "creationTime", "lastUpdated"),
		flags.WithPipelineAliases("creationTimeDateFrom", "time", "creationTime", "lastUpdated"),
		flags.WithPipelineAliases("group", "source.id", "managedObject.id", "id"),

		flags.WithCollectionProperty(".managedObjects[]"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	// Runtime flag options
	flags.WithOptions(
		cmd,
		flags.WithRuntimePipelineProperty(),
	)
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
		flags.WithBoolValue("skipChildrenNames", "skipChildrenNames", ""),
		flags.WithBoolValue("withChildren", "withChildren", ""),
		flags.WithBoolValue("withChildrenCount", "withChildrenCount", ""),
		flags.WithBoolValue("withGroups", "withGroups", ""),
		flags.WithBoolValue("withParents", "withParents", ""),

		flags.WithCumulocityQuery(
			[]flags.GetOption{
				flags.WithStringValue("query", "query", "%s"),
				flags.WithStringValue("name", "name", "(name eq '%s')"),
				flags.WithStringValue("type", "type", "(type eq '%s')"),
				flags.WithDefaultBoolValue("agents", "agents", "has(com_cumulocity_model_Agent)"),
				flags.WithStringValue("fragmentType", "fragmentType", "has(%s)"),
				flags.WithStringValue("owner", "owner", "(owner eq '%s')"),
				flags.WithStringValue("availability", "availability", "(c8y_Availability.status eq '%s')"),
				flags.WithEncodedRelativeTimestamp("lastMessageDateTo", "lastMessageDateTo", "(c8y_Availability.lastMessage le '%s')"),
				flags.WithEncodedRelativeTimestamp("lastMessageDateFrom", "lastMessageDateFrom", "(c8y_Availability.lastMessage ge '%s')"),
				flags.WithEncodedRelativeTimestamp("creationTimeDateTo", "creationTimeDateTo", "(creationTime.date le '%s')"),
				flags.WithEncodedRelativeTimestamp("creationTimeDateFrom", "creationTimeDateFrom", "(creationTime.date ge '%s')"),
				c8yfetcher.WithDeviceGroupByNameFirstMatch(n.factory, args, "group", "group", "bygroupid(%s)"),
			},
			"q",
		),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}
	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParameters(query)

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
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder(false)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "GET",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
