// Prototype of a v2-based list command: the CLI input machinery resolves a
// Cumulocity query expression per input item, and each item drives a typed
// Devices.ListAll call whose pages stream through the shared output pipeline.
package list

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/devices"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// List2Cmd command
type List2Cmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewList2Cmd creates a command to Get device collection
func NewList2Cmd(f *cmdutil.Factory) *List2Cmd {
	ccmd := &List2Cmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list2",
		Short: "Get device collection",
		Long:  `Get a collection of devices based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y devices list2 --name "sensor*" --type myType
Get a collection of devices of type "myType", and their names start with "sensor"

$ c8y devices list2 --query "name eq '*sensor*' and creationTime.date gt '2021-04-02T00:00:00'"
Get devices which names containing 'sensor' and were created after 2021-04-02

$ echo -e "c8y_MacOS\nc8y_Linux" | c8y devices list2 --queryTemplate "type eq '%s'"
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
	cmd.Flags().Bool("withLatestValues", false, "(FEATURE_PREVIEW) Include c8y_LatestMeasurements fragment, which contains the latest measurement values reported by the device to the platform")

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

		flags.WithCollectionProperty("managedObjects"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *List2Cmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	// Cumulocity query expression. The template machinery handles pipeline
	// input (--query accepts pipeline, --queryTemplate formatting) and the
	// group name -> id lookup, producing one "q" value per input item.
	//
	// Note: custom query parameters (cfg.GetQueryParameters()) are not wired
	// up here: typed service options have no escape hatch for arbitrary
	// parameters yet. See the request-modifier seam in the v2 integration
	// design notes.
	err = r.Bind(c8ystream.Query(
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
	))
	if err != nil {
		return err
	}

	// Typed options shared by every request; the per-item query expression is
	// filled in from Args.
	baseOpt := devices.ListOptions{}
	baseOpt.SkipChildrenNames, _ = cmd.Flags().GetBool("skipChildrenNames")
	baseOpt.WithChildren, _ = cmd.Flags().GetBool("withChildren")
	baseOpt.WithChildrenCount, _ = cmd.Flags().GetBool("withChildrenCount")
	baseOpt.WithParents, _ = cmd.Flags().GetBool("withParents")
	baseOpt.WithLatestValues, _ = cmd.Flags().GetBool("withLatestValues")
	baseOpt.PaginationOptions = pagination.PaginationOptions{
		PageSize:          r.Config.GetPageSize(),
		WithTotalPages:    r.Config.WithTotalPages(),
		WithTotalElements: r.Config.WithTotalElements(),
		CurrentPage:       int(r.Config.GetCurrentPage()),
		MaxItems:          r.Config.MaxItems(),
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(ctx context.Context, args c8ystream.Args) output.Seq {
		opt := baseOpt
		opt.Query = args.Query("q")
		return output.FromIterator(client.Devices.ListAll(ctx, opt).Items())
	})
}
