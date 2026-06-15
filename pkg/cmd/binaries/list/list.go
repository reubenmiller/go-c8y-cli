// v2-based binary list: lists the metadata of inventory binaries via
// Binaries.ListAll. The type flag drives iteration (pipe or --type); the other
// filters fill the typed ListOptions. childDeviceId resolves a device name to
// its id.
package list

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/binaries"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get binary collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get binary collection",
		Long:  `Get a collection of inventory binaries. The results include the meta information about binary and not the binary itself.`,
		Example: heredoc.Doc(`
$ c8y binaries list --pageSize 100
Get a list of binaries

$ c8y binaries list --type package_debian
Get a list of binaries with the type package_debian
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("ids", []string{""}, "The managed object IDs to search for.")
	cmd.Flags().String("type", "", "The type of managed object to search for. (accepts pipeline)")
	cmd.Flags().String("owner", "", "Username of the owner of the managed objects.")
	cmd.Flags().String("text", "", "Search for managed objects where any property value is equal to the given one. Only string values are supported.")
	cmd.Flags().String("childAdditionId", "", "Search for a specific child addition and list all the groups to which it belongs.")
	cmd.Flags().String("childAssetId", "", "Search for a specific child asset and list all the groups to which it belongs.")
	cmd.Flags().StringSlice("childDeviceId", []string{""}, "Search for a specific child device and list all the groups to which it belongs.")

	completion.WithOptions(
		cmd,
		completion.WithDevice("childDeviceId", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("type", "type", false, "type"),
		flags.WithCollectionProperty("managedObjects"),
		flags.WithPowershellName("Get-BinaryCollection"),
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
		opt := binaries.ListOptions{
			Type:            in.String("type"),
			Owner:           in.String("owner"),
			ChildAdditionID: in.String("childAdditionId"),
			ChildAssetID:    in.String("childAssetId"),
		}
		for _, id := range in.StringSlice("ids") {
			if id != "" {
				opt.Ids = append(opt.Ids, id)
			}
		}
		// childDeviceId is a device reference: resolve a name to its id.
		for _, device := range in.StringSlice("childDeviceId") {
			if device != "" {
				id, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(device), nil)
				if err != nil {
					return nil, err
				}
				opt.ChildDeviceID = id
				break
			}
		}
		// GAP: --text has no field in the generated binaries ListOptions.
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.Binaries.ListAll), nil
	})
}
