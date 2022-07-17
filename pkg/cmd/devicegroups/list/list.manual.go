package list

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdList struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdList(f *cmdutil.Factory) *CmdList {
	ccmd := &CmdList{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get device group collection",
		Long:  `Get a collection of device groups based on filter parameters`,
		Example: heredoc.Doc(`
		c8y devicesgroups list --name "MyGroup*"

		Get a collection of device groups with names that start with "MyGroup"
		`),
		RunE: ccmd.getDeviceGroupCollection,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("type", "", "Filter by type")
	cmd.Flags().String("fragmentType", "", "Filter by fragment type")
	cmd.Flags().String("owner", "", "Filter by owner")
	cmd.Flags().String("query", "", "Additional query filter (accepts pipeline)")
	cmd.Flags().String("queryTemplate", "", "String template to be used when applying the given query. Use %s to reference the query/pipeline input")
	cmd.Flags().Bool("excludeRootGroup", false, "Exclude root groups from the list")
	cmd.Flags().Bool("withParents", false, "Include a flat list of all parents and grandparents of the given object")
	cmd.Flags().Bool("withChildren", false, "Include names of child assets (only use where necessary as it is slow for large groups)")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("query", "query", false),
	)

	// Required flags
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdList) getDeviceGroupCollection(cmd *cobra.Command, args []string) error {
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
		flags.WithStaticStringValue("devicegroup", "(has(c8y_IsDeviceGroup))"),
		flags.WithStringValue("name", "name", "(name eq '%s')"),
		flags.WithStringValue("type", "type", "(type eq '%s')"),
		flags.WithStringValue("fragmentType", "fragmentType", "has(%s)"),
		flags.WithStringValue("owner", "owner", "(owner eq '%s')"),
		flags.WithBoolValue("excludeRootGroup", "excludeRootGroup", "not(type eq 'c8y_DeviceGroup')"),
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
		flags.WithDefaultBoolValue("withChildren", "withChildren"),
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
		flags.WithCustomStringValue(
			flags.BuildCumulocityQuery(cmd, c8yQueryParts, orderBy),
			func() string { return "query" },
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
	body := mapbuilder.NewInitializedMapBuilder(false)

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects")

	req := c8y.RequestOptions{
		Method:       "GET",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
