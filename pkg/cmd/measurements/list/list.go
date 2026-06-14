// v2-based measurements list: fills the typed measurements.ListOptions from the
// flags; the source device is resolved by go-c8y; time-keyset pagination.
//
// GAP: the legacy --csvFormat/--excelFormat/--unit options are not carried over
// — the typed measurements.ListOptions does not expose them (they change the
// Accept header / add a query param). They need a per-call request-modifier seam
// on go-c8y; see proposals/CLI_CODEGEN_INVERSION.md. The deprecated --fragmentType
// is dropped (use --valueFragmentType).
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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/measurements"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get measurement collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get measurement collection",
		Long:  `Get a collection of measurements based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y measurements list --device 12345 --valueFragmentType c8y_Temperature
Get temperature measurements for a device

$ c8y devices list --type myType | c8y measurements list --dateFrom -1d
List the last day of measurements for each piped device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Source device id (accepts pipeline)")
	cmd.Flags().String("type", "", "Measurement type")
	cmd.Flags().String("valueFragmentType", "", "Value fragment type, e.g. c8y_Temperature")
	cmd.Flags().String("valueFragmentSeries", "", "Value fragment series, e.g. T")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of measurement occurrence")
	cmd.Flags().String("dateTo", "", "End date or date and time of measurement occurrence")
	cmd.Flags().Bool("revert", false, "Return the newest instead of the oldest measurements. Must be used with dateFrom and dateTo parameters")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "source", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("measurements"),
		flags.WithPowershellName("Get-MeasurementCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.measurementCollection+json", "application/vnd.com.nsn.cumulocity.measurement+json"),
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
	paginationStrategy := pagination.StrategyKind(r.Config.PaginationStrategy())

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := measurements.ListOptions{
			Type:                in.String("type"),
			ValueFragmentType:   in.String("valueFragmentType"),
			ValueFragmentSeries: in.String("valueFragmentSeries"),
			DateFrom:            in.TimeValue("dateFrom"),
			DateTo:              in.TimeValue("dateTo"),
			Revert:              in.Bool("revert"),
		}
		if device := in.String("device"); device != "" {
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
		return c8ystream.ListCall(rawOutput, opt, client.Measurements.ListAll), nil
	})
}
