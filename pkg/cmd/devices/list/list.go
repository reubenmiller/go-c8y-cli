// v2-based device list command: the "query" flag is the iterating input
// (piped lines feed it, or its own value drives a single run), each item
// builds a Cumulocity inventory query and drives a typed Devices.ListAll call
// whose pages stream through the shared output pipeline.
package list

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/devicegroups"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/devices"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
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

$ echo -e "c8y_MacOS\nc8y_Linux" | c8y devices list --queryTemplate "type eq '%s'"
Get devices with type 'c8y_MacOS' then devices with type 'c8y_Linux' (using pipeline)

$ echo '{"name":"foo","type":"bar"}' | c8y devices list --name -.name --type -.type
Bind several filters to properties of the same piped object
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
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	// The query flag is the iterating input: piped lines feed it (extracting
	// query/c8y_DeviceQueryString from a JSON object, or using a plain line as
	// is), or a single --query value drives one run. Other flags resolve from
	// the same piped object via -.path references.
	if err := r.InputFlag("query"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	// Pagination follows the v1 rules: page size is only sent when it differs
	// from the server default, total page/element counts and the current page
	// only when explicitly requested, and --includeAll raises the page size.
	common, err := r.Config.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	queryTemplate := cmd.Flag("queryTemplate").Value.String()
	orderBy := cmd.Flag("orderBy").Value.String()

	// --raw emits per-page envelopes; otherwise the extracted items. Dry-run
	// is handled by the transport, so it needs no branch here.
	rawOutput := r.Config.RawOutput()
	includeAll := r.Config.IncludeAll()

	// Pagination strategy: auto by default. A full walk (--includeAll) defaults
	// to the SDK's _id keyset — which advances an "_id gt '<last id>'" cursor
	// instead of deep offset paging — by dropping the default name ordering so
	// the keyset applies. An explicit --orderBy or --paginationStrategy wins.
	paginationStrategy := pagination.StrategyKind(r.Config.PaginationStrategy())
	if paginationStrategy == "" {
		paginationStrategy = pagination.StrategyAuto // unset == auto
	}
	if paginationStrategy == pagination.StrategyAuto && includeAll && !cmd.Flags().Changed("orderBy") {
		paginationStrategy = pagination.StrategyIDKeyset
	}
	orderByValue := orderBy
	if paginationStrategy == pagination.StrategyIDKeyset {
		orderByValue = "" // the SDK forces "$orderby=_id asc"; a name order would conflict
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		// Per item: read each flag as a plain value (with -.path input refs) and
		// assemble the Cumulocity inventory query via the SDK builder, which skips
		// empty values, wraps each filter part in parentheses, joins them with
		// " and " and appends the order clause in Build().
		q := model.NewInventoryQuery()

		if raw := in.String("query"); raw != "" {
			q.AddFilterPart(applyQueryTemplate(queryTemplate, raw))
		} else if queryTemplate != "" {
			q.AddFilterPart(queryTemplate)
		}

		q.AddFilterEqStr("name", in.String("name")).
			AddFilterEqStr("type", in.String("type")).
			AddFilter(model.FilterHasFragment("com_cumulocity_model_Agent", in.Bool("agents"))).
			HasFragment(in.String("fragmentType")).
			AddFilterEqStr("owner", in.String("owner")).
			AddFilterEqStr("c8y_Availability.status", in.String("availability")).
			AddFilterOp("c8y_Availability.lastMessage", "le", in.Time("lastMessageDateTo")).
			AddFilterOp("c8y_Availability.lastMessage", "ge", in.Time("lastMessageDateFrom")).
			AddFilterOp("creationTime.date", "le", in.Time("creationTimeDateTo")).
			AddFilterOp("creationTime.date", "ge", in.Time("creationTimeDateFrom"))

		for _, group := range in.StringSlice("group") {
			if group != "" {
				id, err := client.DeviceGroups.ResolveID(in.ResolveContext(), groupRef(group), nil)
				if err != nil {
					return nil, err
				}
				q.ByGroupID(id)
			}
		}
		q.AddOrderBy(orderByValue)

		// The _id keyset cursor is now applied inside the SDK's ListAll (driven by
		// the strategy below); the command only builds the filter and order.
		opt := devices.ListOptions{Query: q.Build()}
		opt.SkipChildrenNames = in.Bool("skipChildrenNames")
		opt.WithChildren = in.Bool("withChildren")
		opt.WithChildrenCount = in.Bool("withChildrenCount")
		opt.WithParents = in.Bool("withParents")
		opt.WithLatestValues = in.Bool("withLatestValues")
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		// Shared --raw / default (and transparent dry-run) routing from a
		// single paginating method, identical across all list commands.
		return c8ystream.ListCall(rawOutput, opt, client.Devices.ListAll), nil
	})
}

// applyQueryTemplate formats a raw query value with the --queryTemplate (a
// %s template). An empty template returns the value unchanged.
func applyQueryTemplate(template, value string) string {
	if template == "" {
		return value
	}
	return fmt.Sprintf(template, value)
}

// groupRef builds a device-group resolver reference: an all-digit value is an
// id (used as-is), anything else is treated as a name to look up.
func groupRef(value string) string {
	for _, r := range value {
		if r < '0' || r > '9' {
			return string(devicegroups.ByName(value))
		}
	}
	return value
}
