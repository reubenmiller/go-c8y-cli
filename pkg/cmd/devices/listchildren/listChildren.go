// v2-based (deprecated) list child devices: streams a device's childDevice
// references via ManagedObjects.ChildDevices.ListAll, plucking the nested managed
// objects. The device flag drives iteration. Hidden and deprecated in favour of
// `c8y devices children list --childType device`.
package listchildren

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects/childdevices"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/spf13/cobra"
)

// ListChildrenCmd command
type ListChildrenCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListChildrenCmd creates a command to Get child device collection
func NewListChildrenCmd(f *cmdutil.Factory) *ListChildrenCmd {
	ccmd := &ListChildrenCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "listChildren",
		Short:  "Get child device collection",
		Long:   `Get a collection of child managedObject references`,
		Hidden: true,
		Example: heredoc.Doc(`
$ c8y devices listChildren --device 12345
Get a list of the child devices of an existing device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device. (required) (accepts pipeline)")
	cmd.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "device", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("references.#.managedObject"),
		flags.WithPowershellName("Get-ChildDeviceCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObjectReferenceCollection+json", "application/vnd.com.nsn.cumulocity.managedObject+json"),
		flags.WithDeprecationNotice("please use 'c8y devices children list --childType device' instead"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListChildrenCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("device"); err != nil {
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
		parentID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("device")), nil)
		if err != nil {
			return nil, err
		}

		opt := childdevices.ListOptions{}
		opt.WithChildren = in.Bool("withChildren")
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
		}

		listAll := func(ctx context.Context, o childdevices.ListOptions) *pagination.Iterator[jsonmodels.ManagedObject] {
			return client.ManagedObjects.ChildDevices.ListAll(ctx, parentID, o)
		}
		return c8ystream.ListCall(rawOutput, opt, listAll), nil
	})
}
