package list

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdAgentList struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdAgentList(f *cmdutil.Factory) *CmdAgentList {
	ccmd := &CmdAgentList{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get agent collection",
		Long:  `Get a collection of agents based on filter parameters`,
		Example: heredoc.Doc(`
			$ c8y agents list --name "sensor*" --type myType
			Get a collection of agents with type "myType", and their names start with "sensor"

			$ echo "name eq 'sensor*'" | c8y agents list
			Get a collection of agents with names starting with "sensor" using a piped inventory query (or could be piped from a file)

			$ c8y agents list --creationTimeDateFrom -7d
			Get agents which where registered longer than 7 days ago

			$ c8y agents list --creationTimeDateTo -1d
			Get agents which where registered in the last day

			$ c8y agents list --name "my example agent" --select type --output csv | c8y agents list --queryTemplate "type eq '%s'"
			Find an agent by name, then find other agents which the same type
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("type", "", "Filter by type")
	cmd.Flags().String("fragmentType", "", "Filter by fragment type")
	cmd.Flags().String("owner", "", "Filter by owner")
	cmd.Flags().String("availability", "", "Filter by c8y_Availability.status")
	cmd.Flags().String("lastMessageDateTo", "", "Filter c8y_Availability.lastMessage to a specific date")
	cmd.Flags().String("lastMessageDateFrom", "", "Filter c8y_Availability.lastMessage from a specific date")
	cmd.Flags().String("creationTimeDateTo", "", "Filter creationTime.date to a specific date")
	cmd.Flags().String("creationTimeDateFrom", "", "Filter creationTime.date from a specific date")
	cmd.Flags().String("group", "", "Filter by group inclusion")
	cmd.Flags().Bool("withParents", false, "Include a flat list of all parents and grandparents of the given object")

	flags.WithOptions(
		cmd,
		flags.WithCommonCumulocityQueryOptions(),
		flags.WithExtendedPipelineSupport("query", "query", false, "c8y_DeviceQueryString"),
	)

	// Required flags
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdAgentList) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
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

	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	commonOptions.ResultProperty = "managedObjects"
	commonOptions.AddQueryParameters(query)

	c8yQueryParts, err := flags.WithC8YQueryOptions(
		cmd,
		inputIterators,
		flags.WithStaticStringValue("agent", "(has(com_cumulocity_model_Agent))"),
		flags.WithStringValue("name", "name", "(name eq '%s')"),
		flags.WithStringValue("type", "type", "(type eq '%s')"),
		flags.WithStringValue("fragmentType", "fragmentType", "has(%s)"),
		flags.WithStringValue("owner", "owner", "(owner eq '%s')"),
		c8yfetcher.WithDeviceGroupByNameFirstMatch(client, args, "group", "group", "bygroupid(%s)"),
		flags.WithStringValue("availability", "availability", "(c8y_Availability.status eq '%s')"),
		flags.WithRelativeTimestamp("lastMessageDateTo", "lastMessageDateTo", "(c8y_Availability.lastMessage le '%s')"),
		flags.WithRelativeTimestamp("lastMessageDateFrom", "lastMessageDateFrom", "(c8y_Availability.lastMessage ge '%s')"),
		flags.WithRelativeTimestamp("creationTimeDateTo", "creationTimeDateTo", "creationTime.date le '%s'"),
		flags.WithRelativeTimestamp("creationTimeDateFrom", "creationTimeDateFrom", "creationTime.date ge '%s'"),
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
		Header:       headers,
		DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
