// v2-based inventory list: the type flag is the iterating input (piped lines feed
// it, or its own value drives a single run); each item builds a managed-object
// list filter and drives a typed ManagedObjects.ListAll call whose pages stream
// through the shared output pipeline.
package list

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get managed object collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get managed object collection",
		Long:  `Get a collection of managedObjects based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y inventory list
Get a list of managed objects

$ c8y inventory list --ids 1111,2222
Get a list of managed objects by ids

$ echo 'myType' | c8y inventory list
Search by type using pipeline. piped input will be mapped to type parameter

$ c8y inventory get --id 1234 | c8y inventory list
Get managed objects which have the same type as the managed object id=1234. piped input will be mapped to type parameter
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("ids", []string{""}, "List of ids.")
	cmd.Flags().String("type", "", "ManagedObject type. (accepts pipeline)")
	cmd.Flags().String("fragmentType", "", "ManagedObject fragment type.")
	cmd.Flags().String("owner", "", "List of managed objects that are owned by the given username.")
	cmd.Flags().String("text", "", "Search for managed objects where a property value is equal to the given one. The following properties are examined: id, type, name, owner, externalIds.")
	cmd.Flags().Bool("onlyRoots", false, "When set to `true` it returns managed objects which don't have any parent. If the current user doesn't have access to the parent, this is also root for the user")
	cmd.Flags().String("childAdditionId", "", "Search for a specific child addition and list all the groups to which it belongs.")
	cmd.Flags().String("childAssetId", "", "Search for a specific child asset and list all the groups to which it belongs.")
	cmd.Flags().StringSlice("childDeviceId", []string{""}, "Search for a specific child device and list all the groups to which it belongs.")
	cmd.Flags().Bool("skipChildrenNames", false, "Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved")
	cmd.Flags().Bool("withParents", false, "Include a flat list of all parents and grandparents of the given object")
	cmd.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.")
	cmd.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")
	cmd.Flags().Bool("withGroups", false, "When set to true it returns additional information about the groups to which the searched managed object belongs. This results in setting the assetParents property with additional information about the groups.")
	cmd.Flags().Bool("withLatestValues", false, "(FEATURE_PREVIEW) Include c8y_LatestMeasurements fragment, which contains the latest measurement values reported by the device to the platform")

	completion.WithOptions(
		cmd,
		completion.WithDevice("childDeviceId", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("type", "type", false, "type"),
		flags.WithPipelineAliases("childDeviceId", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("managedObjects"),
		flags.WithPowershellName("Get-ManagedObjectCollection"),
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

	if err := r.InputFlag("type"); err != nil {
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

	paginationStrategy := pagination.StrategyKind(r.Config.PaginationStrategy())

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		childDeviceID := firstNonEmpty(in.StringSlice("childDeviceId"))
		if childDeviceID != "" {
			childDeviceID, err = client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(childDeviceID), nil)
			if err != nil {
				return nil, err
			}
		}

		opt := managedobjects.ListOptions{
			Ids:             nonEmpty(in.StringSlice("ids")),
			Type:            in.String("type"),
			FragmentType:    in.String("fragmentType"),
			Owner:           in.String("owner"),
			Text:            in.String("text"),
			OnlyRoots:       in.Bool("onlyRoots"),
			ChildAdditionID: in.String("childAdditionId"),
			ChildAssetID:    in.String("childAssetId"),
			ChildDeviceID:   childDeviceID,
		}
		opt.SkipChildrenNames = in.Bool("skipChildrenNames")
		opt.WithParents = in.Bool("withParents")
		opt.WithChildren = in.Bool("withChildren")
		opt.WithChildrenCount = in.Bool("withChildrenCount")
		opt.WithGroups = in.Bool("withGroups")
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

// firstNonEmpty returns the first non-empty entry of a slice flag.
func firstNonEmpty(values []string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// nonEmpty drops empty entries from a slice flag (the StringSlice default is
// `[""]`, which would otherwise serialise as an empty ids filter).
func nonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
