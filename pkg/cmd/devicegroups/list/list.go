// v2-based device group list: the "query" flag is the iterating input (piped
// lines feed it, or its own value drives a single run); each item builds a
// Cumulocity inventory query — always scoped to device groups
// (has(c8y_IsDeviceGroup)) — and drives a typed DeviceGroups.ListAll call whose
// pages stream through the shared output pipeline.
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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get device group collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get device group collection",
		Long:  `Get a collection of device groups based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y devicegroups list --name "parent*"
Get a collection of device groups with names that start with 'parent'
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("queryTemplate", "", "String template to be used when applying the given query. Use %s to reference the query/pipeline input")
	cmd.Flags().String("orderBy", "name", "Order by. e.g. _id asc or name asc or creationTime.date desc")
	cmd.Flags().String("query", "", "Additional query filter (accepts pipeline)")
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("type", "", "Filter by type")
	cmd.Flags().String("fragmentType", "", "Filter by fragment type")
	cmd.Flags().String("owner", "", "Filter by owner")
	cmd.Flags().Bool("excludeRootGroup", false, "Filter by group inclusion")
	cmd.Flags().StringSlice("group", []string{""}, "Filter by group inclusion")
	cmd.Flags().Bool("skipChildrenNames", false, "Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved")
	cmd.Flags().Bool("withChildren", false, "Include names of child assets (only use where necessary as it is slow for large groups)")
	cmd.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")
	cmd.Flags().Bool("withGroups", false, "When set to true it returns additional information about the groups to which the searched managed object belongs. This results in setting the assetParents property with additional information about the groups.")
	cmd.Flags().Bool("withParents", false, "Include a flat list of all parents and grandparents of the given object")

	completion.WithOptions(
		cmd,
		completion.WithDeviceGroup("group", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("query", "query", false, "c8y_DeviceQueryString"),
		flags.WithCollectionProperty("managedObjects"),
		flags.WithPowershellName("Get-DeviceGroupCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedobjectcollection+json", "application/vnd.com.nsn.cumulocity.customDeviceGroup+json"),
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

		// Always scope to device groups (fragment c8y_IsDeviceGroup); a plain
		// inventory query would also return ordinary devices.
		q.HasFragment(devicegroups.FragmentIsDeviceGroup).
			AddFilterEqStr("name", in.String("name")).
			AddFilterEqStr("type", in.String("type")).
			HasFragment(in.String("fragmentType")).
			AddFilterEqStr("owner", in.String("owner"))

		if in.Bool("excludeRootGroup") {
			q.AddFilterPart("not(type eq 'c8y_DeviceGroup')")
		}

		// Filter by parent group(s): resolve each reference (id or name) and add a
		// bygroupid term. Resolution hits the real API even under --dry.
		for _, g := range in.StringSlice("group") {
			if g == "" {
				continue
			}
			id, err := client.DeviceGroups.ResolveID(in.ResolveContext(), c8ystream.NameOrID(g), nil)
			if err != nil {
				return nil, err
			}
			q.ByGroupID(id)
		}

		q.AddOrderBy(orderByValue)

		opt := devicegroups.ListOptions{Query: q.Build()}
		opt.SkipChildrenNames = in.Bool("skipChildrenNames")
		opt.WithChildren = in.Bool("withChildren")
		opt.WithChildrenCount = in.Bool("withChildrenCount")
		opt.WithGroups = in.Bool("withGroups")
		opt.WithParents = in.Bool("withParents")
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.DeviceGroups.ListAll), nil
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
