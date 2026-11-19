// v2-based inventory findByText: the text flag is the iterating input (required;
// piped lines feed it); each item drives a typed ManagedObjects.ListAll call
// filtered by the Cumulocity text-search algorithm.
package findbytext

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// FindByTextCmd command
type FindByTextCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewFindByTextCmd creates a command to Find managed object by text collection
func NewFindByTextCmd(f *cmdutil.Factory) *FindByTextCmd {
	ccmd := &FindByTextCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "findByText",
		Short: "Find managed object by text collection",
		Long:  `Find a collection of managedObjects which match a given text value`,
		Example: heredoc.Doc(`
$ c8y inventory findByText --text "myText"
Get a list of managed objects

$ echo "myText" | c8y inventory findByText
Find managed objects which contain the text 'myText' (using pipeline)

$ echo "myText" | c8y inventory findByText --fragmentType c8y_IsDevice
Find managed objects which contain the text 'myText' and is a device (using pipeline)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("text", "", "Search for managed objects where a property value is equal to the given one. The following properties are examined: id, type, name, owner, externalIds. (required) (accepts pipeline)")
	cmd.Flags().String("type", "", "ManagedObject type.")
	cmd.Flags().String("fragmentType", "", "ManagedObject fragment type.")
	cmd.Flags().Bool("skipChildrenNames", false, "Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved")
	cmd.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.")
	cmd.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")
	cmd.Flags().Bool("withGroups", false, "When set to true it returns additional information about the groups to which the searched managed object belongs. This results in setting the assetParents property with additional information about the groups.")
	cmd.Flags().Bool("withParents", false, "Include a flat list of all parents and grandparents of the given object")
	cmd.Flags().Bool("withLatestValues", false, "(FEATURE_PREVIEW) Include c8y_LatestMeasurements fragment, which contains the latest measurement values reported by the device to the platform")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("text", "text", true, "id"),
		flags.WithCollectionProperty("managedObjects"),
		flags.WithPowershellName("Find-ByTextManagedObjectCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObjectCollection+json", "application/vnd.com.nsn.cumulocity.managedObject+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *FindByTextCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("text"); err != nil {
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
		opt := managedobjects.ListOptions{
			Text:         in.String("text"),
			Type:         in.String("type"),
			FragmentType: in.String("fragmentType"),
		}
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
