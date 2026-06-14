// v2-based alarms deleteCollection: deletes the set of alarms matching the given
// filter (device/date/status/severity) in one call via Alarms.DeleteList. The
// device flag drives iteration (pipe or --device). Success yields no output.
package deletecollection

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/alarms"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteCollectionCmd command
type DeleteCollectionCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteCollectionCmd creates a command to Delete alarm collection
func NewDeleteCollectionCmd(f *cmdutil.Factory) *DeleteCollectionCmd {
	ccmd := &DeleteCollectionCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "deleteCollection",
		Short: "Delete alarm collection",
		Long:  `Delete a collection of alarms by a given filter`,
		Example: heredoc.Doc(`
$ c8y alarms deleteCollection --device 12345 --severity MAJOR
Remove alarms on the device with the severity set to MAJOR

$ c8y alarms deleteCollection --device 12345 --dateFrom "-10m" --status ACTIVE
Remove alarms on the device which are active and created in the last 10 minutes
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Source device id. (accepts pipeline)")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of alarm occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of alarm occurrence.")
	cmd.Flags().String("createdFrom", "", "Start date or date and time of the alarm creation.")
	cmd.Flags().String("createdTo", "", "End date or date and time of the alarm creation.")
	cmd.Flags().String("type", "", "Alarm type.")
	cmd.Flags().StringSlice("status", []string{""}, "Comma separated alarm statuses, for example ACTIVE,CLEARED.")
	cmd.Flags().String("severity", "", "Alarm severity, for example CRITICAL, MAJOR, MINOR or WARNING.")
	cmd.Flags().Bool("resolved", false, "When set to true only resolved alarms will be removed (the one with status CLEARED), false means alarms with status ACTIVE or ACKNOWLEDGED.")
	cmd.Flags().Bool("withSourceAssets", false, "When set to true also alarms for related source assets will be removed. When this parameter is provided also source must be defined.")
	cmd.Flags().Bool("withSourceDevices", false, "When set to true also alarms for related source devices will be removed. When this parameter is provided also source must be defined.")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("severity", "CRITICAL", "MAJOR", "MINOR", "WARNING"),
		completion.WithValidateSet("status", "ACTIVE", "ACKNOWLEDGED", "CLEARED"),
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("device", "source.id", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Remove-AlarmCollection"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
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

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := alarms.DeleteListOptions{
			DateFrom:          in.TimeValue("dateFrom"),
			DateTo:            in.TimeValue("dateTo"),
			CreatedFrom:       in.TimeValue("createdFrom"),
			CreatedTo:         in.TimeValue("createdTo"),
			Resolved:          in.Bool("resolved"),
			WithSourceAssets:  in.Bool("withSourceAssets"),
			WithSourceDevices: in.Bool("withSourceDevices"),
		}
		if t := in.String("type"); t != "" {
			opt.Type = []string{t}
		}
		if s := in.String("severity"); s != "" {
			opt.Severity = []model.AlarmSeverity{model.AlarmSeverity(s)}
		}
		for _, status := range in.StringSlice("status") {
			if status != "" {
				opt.Status = append(opt.Status, model.AlarmStatus(status))
			}
		}
		if device := in.String("device"); device != "" {
			opt.Source = managedobjects.DeviceRef(c8ystream.NameOrID(device))
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Alarms.DeleteList(ctx, opt)
			})
		}, nil
	})
}
