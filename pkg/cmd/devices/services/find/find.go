// v2-based find device services: a single global inventory query for c8y_Service
// managed objects (optionally filtered by serviceType/name/status or a raw
// query), streamed via ManagedObjects.ListAll. Run-once — there is no driving
// input flag.
package find

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/services/serviceref"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
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

// NewFindCmd creates a command to Find services
func NewFindCmd(f *cmdutil.Factory) *FindCmd {
	ccmd := &FindCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find services",
		Long:  `Find services of any device`,
		Example: heredoc.Doc(`
$ c8y devices services find
Find all services (from any device)

$ c8y devices services find --status down
Find any services which are currently down

$ c8y devices services find --name ntp --status down
Find any ntp services which are currently down
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("query", "", "Additional query filter")
	cmd.Flags().String("serviceType", "", "Filter by service type e.g. systemd")
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("status", "", "Filter by service status (custom values allowed)")
	cmd.Flags().String("orderBy", "", "Order by. e.g. _id asc or name asc or creationTime.date desc")
	cmd.Flags().Bool("skipChildrenNames", false, "Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved")
	cmd.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.")
	cmd.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")
	cmd.Flags().Bool("withGroups", false, "When set to true it returns additional information about the groups to which the searched managed object belongs. This results in setting the assetParents property with additional information about the groups.")
	cmd.Flags().Bool("withParents", false, "Include a flat list of all parents and grandparents of the given object")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "up", "down", "unknown"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("", "", false),
		flags.WithCollectionProperty("managedObjects"),
		flags.WithPowershellName("Find-DeviceServiceCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObjectReferenceCollection+json", "application/vnd.com.nsn.cumulocity.managedObject+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *FindCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
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
	includeAll := r.Config.IncludeAll()

	// Pagination strategy mirrors devices list: auto by default; a full walk
	// (--includeAll) defaults to the SDK's _id keyset (dropping any order so the
	// keyset applies) unless --orderBy/--paginationStrategy is set.
	paginationStrategy := pagination.StrategyKind(r.Config.PaginationStrategy())
	if paginationStrategy == "" {
		paginationStrategy = pagination.StrategyAuto
	}
	if paginationStrategy == pagination.StrategyAuto && includeAll && !cmd.Flags().Changed("orderBy") {
		paginationStrategy = pagination.StrategyIDKeyset
	}
	orderByValue := cmd.Flag("orderBy").Value.String()
	if paginationStrategy == pagination.StrategyIDKeyset {
		orderByValue = ""
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		q := model.NewInventoryQuery()
		q.AddFilterEqStr("type", serviceref.ServiceType)
		if raw := in.String("query"); raw != "" {
			q.AddFilterPart(raw)
		}
		q.AddFilterEqStr("serviceType", in.String("serviceType")).
			AddFilterEqStr("name", in.String("name")).
			AddFilterEqStr("status", in.String("status"))
		q.AddOrderBy(orderByValue)

		opt := managedobjects.ListOptions{Query: q.Build()}
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
		return c8ystream.ListCall(rawOutput, opt, client.ManagedObjects.ListAll), nil
	})
}
