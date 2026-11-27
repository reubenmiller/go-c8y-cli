// v2-based measurements list: fills the typed measurements.ListOptions from the
// flags; the source device is resolved by go-c8y; time-keyset pagination.
// The legacy header parameters are carried via the request context (--unit →
// X-Cumulocity-System-Of-Units) or a dedicated export call (--csvFormat /
// --excelFormat → Accept-driven single-document download, no pagination); the
// deprecated --fragmentType rides along as an extra query parameter.

package measurements

import (
	"context"
	"net/http"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/measurements"
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
	cmd.Flags().String("fragmentType", "", "Fragment name from measurement (deprecated)")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of measurement occurrence")
	cmd.Flags().String("dateTo", "", "End date or date and time of measurement occurrence")
	cmd.Flags().Bool("revert", false, "Return the newest instead of the oldest measurements. Must be used with dateFrom and dateTo parameters")
	cmd.Flags().Bool("csvFormat", false, "Results will be displayed in csv format. Note: -IncludeAll, is not supported when using using this parameter")
	cmd.Flags().Bool("excelFormat", false, "Results will be displayed in Excel format Note: -IncludeAll, is not supported when using using this parameter")
	cmd.Flags().String("unit", "", "Every measurement fragment which contains 'unit' property will be transformed to use required system of units.")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("unit", "imperial", "metric"),
		completion.WithDevice("device", ccmd.factory.ClientV2),
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

	paging, err := r.PaginationByPage()
	if err != nil {
		return err
	}

	// v1 header parameters: --csvFormat/--excelFormat switch the Accept header
	// (excel wins when both are given, matching the v1 getter order) and --unit
	// selects the system of units.
	exportAccept := ""
	if in, _ := cmd.Flags().GetBool("csvFormat"); in {
		exportAccept = "text/csv"
	}
	if in, _ := cmd.Flags().GetBool("excelFormat"); in {
		exportAccept = "application/vnd.ms-excel"
	}

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
		opt.PaginationOptions = paging

		unit := in.String("unit")
		fragmentType := in.String("fragmentType")
		withParams := func(ctx context.Context) context.Context {
			if !cmd.Flags().Changed("revert") {
				// the SDK always sends revert=false; v2.59.0 sent revert only when the flag is given
				ctx = apiv2.WithRequestHook(ctx, dropRevertFalse)
			}
			ctx = c8ystream.WithHeaderParam(ctx, "X-Cumulocity-System-Of-Units", unit)
			return c8ystream.WithQueryParam(ctx, "fragmentType", fragmentType)
		}

		if exportAccept != "" {
			// CSV/Excel export: the server streams the whole matching collection
			// as one non-JSON document, so bypass the paginated iterator and write
			// the payload to --outputFileRaw (or stdout), yielding nothing — the
			// runner's rawFileStage must not overwrite the file with rendered docs.
			return func(ctx context.Context) any {
				return client.Measurements.Export(withParams(ctx), opt, exportAccept)
			}, nil
		}

		return func(ctx context.Context) any {
			return client.Measurements.ListAll(withParams(ctx), opt)
		}, nil
	})
}

// dropRevertFalse removes the revert=false query parameter of a request.
func dropRevertFalse(req *http.Request) error {
	if req.URL == nil || !strings.Contains(req.URL.RawQuery, "revert=false") {
		return nil
	}
	params := strings.Split(req.URL.RawQuery, "&")
	kept := params[:0]
	for _, p := range params {
		if p != "revert=false" {
			kept = append(kept, p)
		}
	}
	req.URL.RawQuery = strings.Join(kept, "&")
	return nil
}
