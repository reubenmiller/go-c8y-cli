// v2-based events list: per input item a typed events.ListOptions is filled
// directly from the flags (the source device is resolved by go-c8y; time-window
// filters are typed time.Time) and streamed through the SDK's time-keyset
// paginator. Unlike inventory, events need no q-expression builder.
package list

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/events"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get event collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get event collection",
		Long:  `Get a collection of events based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y events list --device 12345 --type myType
Get events for a device of a given type

$ c8y devices list --type myType | c8y events list
List events for each piped device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Source device id (accepts pipeline)")
	cmd.Flags().String("type", "", "Event type")
	cmd.Flags().String("fragmentType", "", "Fragment name from event")
	cmd.Flags().String("fragmentValue", "", "Allows filtering events by the fragment's value, but only when provided together with fragmentType")
	cmd.Flags().String("createdFrom", "", "Start date or date and time of the event's creation (set by the platform during creation)")
	cmd.Flags().String("createdTo", "", "End date or date and time of the event's creation (set by the platform during creation)")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of event occurrence")
	cmd.Flags().String("dateTo", "", "End date or date and time of event occurrence")
	cmd.Flags().String("lastUpdatedFrom", "", "Start date or date and time of the last update made")
	cmd.Flags().String("lastUpdatedTo", "", "End date or date and time of the last update made")
	cmd.Flags().Bool("revert", false, "Return the newest instead of the oldest events. Must be used with dateFrom and dateTo parameters")
	cmd.Flags().Bool("withSourceAssets", false, "When set to true also events for related source assets will be included in the request. When this parameter is provided a source must be specified")
	cmd.Flags().Bool("withSourceDevices", false, "When set to true also events for related source devices will be included in the request. When this parameter is provided a source must be specified")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "source", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("createdFrom", "time", "creationTime", "lastUpdated"),
		flags.WithPipelineAliases("createdTo", "time", "creationTime", "lastUpdated"),
		flags.WithPipelineAliases("dateFrom", "time", "creationTime", "lastUpdated"),
		flags.WithPipelineAliases("dateTo", "time", "creationTime", "lastUpdated"),
		flags.WithPipelineAliases("lastUpdatedFrom", "time", "creationTime", "lastUpdated"),
		flags.WithPipelineAliases("lastUpdatedTo", "time", "creationTime", "lastUpdated"),
		flags.WithCollectionProperty("events"),
		flags.WithPowershellName("Get-EventCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.eventcollection+json", "application/vnd.com.nsn.cumulocity.event+json"),
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

	// device drives iteration: piped device items feed it (extracting
	// source/id/...), a single --device drives one run, and no device lists all
	// events.
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
	// Events page by a time-window keyset inside the SDK; the strategy is passed
	// through (auto by default) and resolved by Events.ListAll.
	paginationStrategy := pagination.StrategyKind(r.Config.PaginationStrategy())

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := events.ListOptions{
			Type:              in.String("type"),
			FragmentType:      in.String("fragmentType"),
			FragmentValue:     in.String("fragmentValue"),
			CreatedFrom:       in.TimeValue("createdFrom"),
			CreatedTo:         in.TimeValue("createdTo"),
			DateFrom:          in.TimeValue("dateFrom"),
			DateTo:            in.TimeValue("dateTo"),
			LastUpdatedFrom:   in.TimeValue("lastUpdatedFrom"),
			LastUpdatedTo:     in.TimeValue("lastUpdatedTo"),
			Revert:            in.Bool("revert"),
			WithSourceAssets:  in.Bool("withSourceAssets"),
			WithSourceDevices: in.Bool("withSourceDevices"),
		}
		if device := in.String("device"); device != "" {
			// go-c8y resolves a device name -> id (plain ids pass through).
			opt.Source = managedobjects.DeviceRef(c8ystream.NameOrID(device))
		}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.Events.ListAll), nil
	})
}
