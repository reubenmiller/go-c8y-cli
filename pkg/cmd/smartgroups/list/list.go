// v2-based smart group list: the "query" flag is the iterating input (piped
// lines feed it, or its own value drives a single run); each item builds a
// Cumulocity inventory query — always scoped to smart groups (type
// c8y_DynamicGroup) — and drives a typed SmartGroups.ListAll call whose pages
// stream through the shared output pipeline.
package list

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/smartgroups"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to List smart group collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List smart group collection",
		Long:  `Get a collection of smart groups based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y smartgroups list --name "my*"
Get a collection of smart groups whose names start with "my"

$ c8y smartgroups list --onlyVisible
Get only the visible smart groups
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
	cmd.Flags().String("deviceQuery", "", "Filter by device query")
	cmd.Flags().String("fragmentType", "", "Filter by fragment type")
	cmd.Flags().String("owner", "", "Filter by owner")
	cmd.Flags().Bool("onlyInvisible", false, "Only include invisible smart groups")
	cmd.Flags().Bool("onlyVisible", false, "Only include visible smart groups")
	cmd.Flags().Bool("skipChildrenNames", false, "Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved")
	cmd.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.")
	cmd.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("query", "query", false, "c8y_DeviceQueryString"),
		flags.WithCollectionProperty("managedObjects"),
		flags.WithPowershellName("Get-SmartGroupCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObjectCollection+json", "application/vnd.com.nsn.cumulocity.managedObject+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
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

	queryTemplate := cmd.Flag("queryTemplate").Value.String()
	orderBy := cmd.Flag("orderBy").Value.String()

	rawOutput := r.Config.RawOutput()
	includeAll := r.Config.IncludeAll()

	paginationStrategy := pagination.StrategyKind(r.Config.PaginationStrategy())
	if paginationStrategy == "" {
		paginationStrategy = pagination.StrategyAuto
	}
	if paginationStrategy == pagination.StrategyAuto && includeAll && !cmd.Flags().Changed("orderBy") {
		paginationStrategy = pagination.StrategyIDKeyset
	}
	orderByValue := orderBy
	if paginationStrategy == pagination.StrategyIDKeyset {
		orderByValue = ""
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		q := model.NewInventoryQuery()

		if raw := in.String("query"); raw != "" {
			q.AddFilterPart(applyQueryTemplate(queryTemplate, raw))
		} else if queryTemplate != "" {
			q.AddFilterPart(queryTemplate)
		}

		// Always scope to smart groups (type c8y_DynamicGroup); a plain inventory
		// query would also return ordinary devices/groups.
		q.AddFilterEqStr("type", smartgroups.TypeDynamicGroup).
			AddFilterEqStr("name", in.String("name")).
			AddFilterEqStr("c8y_DeviceQueryString", in.String("deviceQuery")).
			HasFragment(in.String("fragmentType")).
			AddFilterEqStr("owner", in.String("owner"))

		if in.Bool("onlyInvisible") {
			q.AddFilterPart("has(c8y_IsDynamicGroup.invisible)")
		}
		if in.Bool("onlyVisible") {
			q.AddFilterPart("not(has(c8y_IsDynamicGroup.invisible))")
		}
		q.AddOrderBy(orderByValue)

		opt := smartgroups.ListOptions{Query: q.Build()}
		opt.SkipChildrenNames = in.Bool("skipChildrenNames")
		opt.WithChildren = in.Bool("withChildren")
		opt.WithChildrenCount = in.Bool("withChildrenCount")
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.SmartGroups.ListAll), nil
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
