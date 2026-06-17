// v2-based (deprecated) list child assets: streams a group's childAsset
// references via ManagedObjects.ChildAssets.ListAll, plucking the nested managed
// objects. The id flag drives iteration. Hidden and deprecated in favour of
// `c8y devicegroups children list --childType asset`.
package listassets

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects/childassets"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/spf13/cobra"
)

// ListAssetsCmd command
type ListAssetsCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListAssetsCmd creates a command to Get child asset collection
func NewListAssetsCmd(f *cmdutil.Factory) *ListAssetsCmd {
	ccmd := &ListAssetsCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "listAssets",
		Short:  "Get child asset collection",
		Long:   `Get a collection of child managedObject references`,
		Hidden: true,
		Example: heredoc.Doc(`
$ c8y devicegroups listAssets --id 12345
Get a list of the child assets of an existing device group
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Device Group. (required) (accepts pipeline)")
	cmd.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.")
	cmd.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")

	completion.WithOptions(
		cmd,
		completion.WithDeviceGroup("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithCollectionProperty("references.#.managedObject"),
		flags.WithPowershellName("Get-DeviceGroupChildAssetCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObjectReferenceCollection+json", "application/vnd.com.nsn.cumulocity.managedObject+json"),
		flags.WithDeprecationNotice("please use 'c8y devicegroups children list --childType asset' instead"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListAssetsCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
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

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		parentID, err := client.DeviceGroups.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("id")), nil)
		if err != nil {
			return nil, err
		}

		opt := childassets.ListOptions{}
		opt.WithChildren = in.Bool("withChildren")
		opt.WithChildrenCount = in.Bool("withChildrenCount")
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
		}

		listAll := func(ctx context.Context, o childassets.ListOptions) *pagination.Iterator[jsonmodels.ManagedObject] {
			return client.ManagedObjects.ChildAssets.ListAll(ctx, parentID, o)
		}
		return c8ystream.ListCall(rawOutput, opt, listAll), nil
	})
}
