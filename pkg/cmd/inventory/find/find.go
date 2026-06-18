// v2-based inventory find: the query flag is the iterating input; each item
// assembles a Cumulocity inventory query from the typed filter flags (name, type,
// fragment, availability, date ranges, group inclusion, ...) and drives a typed
// ManagedObjects.ListAll call whose pages stream through the shared pipeline.
package find

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// FindCmd command
type FindCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewFindCmd creates a command to Find managed object collection
func NewFindCmd(f *cmdutil.Factory) *FindCmd {
	ccmd := &FindCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find managed object collection",
		Long:  `Get a collection of managedObjects based on the Cumulocity query language`,
		Example: heredoc.Doc(`
$ c8y inventory find --query "name eq 'roomUpperFloor_*'"
Get a list of managed objects

$ echo "myname" | c8y inventory find --queryTemplate "name eq '*%s*'"
Find managed objects which include myname in their names.

$ echo "name eq 'name'" | c8y inventory find --queryTemplate 'not(%s)'
Invert a given query received via piped input (stdin) by using a template
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("query", "", "ManagedObject query (accepts pipeline)")
	cmd.Flags().String("queryTemplate", "", "String template to be used when applying the given query. Use %s to reference the query/pipeline input")
	cmd.Flags().String("orderBy", "", "Order by. e.g. _id asc or name asc or creationTime.date desc")
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
	cmd.Flags().Bool("onlyDevices", false, "Only include devices (deprecated)")
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
		flags.WithPowershellName("Find-ManagedObjectCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObjectCollection+json", "application/vnd.com.nsn.cumulocity.managedObject+json"),
	)

	_ = cmd.Flags().MarkHidden("onlyDevices")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *FindCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("query"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	common, err := r.Config.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}
	rawOutput := r.Config.RawOutput()

	queryTemplate := cmd.Flag("queryTemplate").Value.String()
	orderBy := cmd.Flag("orderBy").Value.String()

	paginationStrategy := pagination.StrategyKind(r.Config.PaginationStrategy())

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		q := model.NewInventoryQuery()

		if raw := in.String("query"); raw != "" {
			q.AddFilterPart(applyQueryTemplate(queryTemplate, raw))
		} else if queryTemplate != "" {
			q.AddFilterPart(queryTemplate)
		}

		q.AddFilterEqStr("name", in.String("name")).
			AddFilterEqStr("type", in.String("type"))
		if in.Bool("agents") {
			q.AddFilterPart("has(com_cumulocity_model_Agent)")
		}
		q.HasFragment(in.String("fragmentType")).
			AddFilterEqStr("owner", in.String("owner")).
			AddFilterEqStr("c8y_Availability.status", in.String("availability")).
			AddFilterOp("c8y_Availability.lastMessage", "le", in.Time("lastMessageDateTo")).
			AddFilterOp("c8y_Availability.lastMessage", "ge", in.Time("lastMessageDateFrom")).
			AddFilterOp("creationTime.date", "le", in.Time("creationTimeDateTo")).
			AddFilterOp("creationTime.date", "ge", in.Time("creationTimeDateFrom"))

		for _, group := range in.StringSlice("group") {
			if group == "" {
				continue
			}
			groupID, err := client.DeviceGroups.ResolveID(in.ResolveContext(), c8ystream.NameOrID(group), nil)
			if err != nil {
				return nil, err
			}
			q.ByGroupID(groupID)
		}

		if in.Bool("onlyDevices") {
			q.AddFilterPart("has(c8y_IsDevice)")
		}
		q.AddOrderBy(orderBy)

		opt := managedobjects.ListOptions{Query: q.Build()}
		opt.SkipChildrenNames = in.Bool("skipChildrenNames")
		opt.WithChildren = in.Bool("withChildren")
		opt.WithChildrenCount = in.Bool("withChildrenCount")
		opt.WithGroups = in.Bool("withGroups")
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
		return c8ystream.ListCall(rawOutput, opt, client.ManagedObjects.ListAll), nil
	})
}

// applyQueryTemplate formats a raw query value with the --queryTemplate (a %s
// template). An empty template returns the value unchanged.
func applyQueryTemplate(template, value string) string {
	if template == "" {
		return value
	}
	return fmt.Sprintf(template, value)
}
