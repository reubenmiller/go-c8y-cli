// v2-based device services list: the device flag drives iteration; each device's
// c8y_Service child additions (optionally filtered by serviceType/name/status or
// a raw query) are streamed via ManagedObjects.ChildAdditions.ListAll, plucking
// the nested managed objects.
package list

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/services/serviceref"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects/childadditions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get device services collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get device services collection",
		Long:  `Get a collection of services of a device`,
		Example: heredoc.Doc(`
$ c8y devices services list --device 12345
Get services for a specific device

$ c8y devices get --id 12345 | c8y devices services list --name ntp
Get services for a specific device (using pipeline)

$ c8y devices services list --device 12345 --status down
Get services which are currently down for a device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device id (required for name lookup) (required) (accepts pipeline)")
	cmd.Flags().String("query", "", "Additional query filter")
	cmd.Flags().String("serviceType", "", "Filter by service type e.g. systemd")
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("status", "", "Filter by service status")
	cmd.Flags().String("orderBy", "", "Order by. e.g. _id asc or name asc or creationTime.date desc")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("status", "up", "down", "unknown"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "device", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("references.#.managedObject"),
		flags.WithPowershellName("Get-DeviceServiceCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObjectReferenceCollection+json", "application/vnd.com.nsn.cumulocity.managedObject+json"),
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

	orderBy := cmd.Flag("orderBy").Value.String()

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		parentID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("device")), nil)
		if err != nil {
			return nil, err
		}

		q := model.NewInventoryQuery()
		q.AddFilterEqStr("type", serviceref.ServiceType)
		if raw := in.String("query"); raw != "" {
			q.AddFilterPart(raw)
		}
		q.AddFilterEqStr("serviceType", in.String("serviceType")).
			AddFilterEqStr("name", in.String("name")).
			AddFilterEqStr("status", in.String("status"))
		q.AddOrderBy(orderBy)

		opt := childadditions.ListOptions{Query: q.Build()}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
		}

		listAll := func(ctx context.Context, o childadditions.ListOptions) *pagination.Iterator[jsonmodels.ManagedObject] {
			return client.ManagedObjects.ChildAdditions.ListAll(ctx, parentID, o)
		}
		return c8ystream.ListCall(rawOutput, opt, listAll), nil
	})
}
