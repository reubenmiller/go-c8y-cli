// v2-based alarms updateCollection: updates the status of the set of alarms
// matching the given filter (device/date/status/severity) in one call via
// Alarms.UpdateList. The device flag drives iteration (pipe or --device). Only
// the alarm status can be changed (newStatus).
package updatecollection

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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// UpdateCollectionCmd command
type UpdateCollectionCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUpdateCollectionCmd creates a command to Update alarm collection
func NewUpdateCollectionCmd(f *cmdutil.Factory) *UpdateCollectionCmd {
	ccmd := &UpdateCollectionCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "updateCollection",
		Short: "Update alarm collection",
		Long:  `Update the status of a collection of alarms by using a filter. Currently only the status of alarms can be changed`,
		Example: heredoc.Doc(`
$ c8y alarms updateCollection --device 12345 --status ACTIVE --newStatus ACKNOWLEDGED
Update the status of all active alarms on a device to ACKNOWLEDGED
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "The ManagedObject that the alarm originated from (accepts pipeline)")
	cmd.Flags().String("status", "", "The status of the alarm to filter by: ACTIVE, ACKNOWLEDGED or CLEARED. Must be upper-case.")
	cmd.Flags().String("severity", "", "The severity of the alarm to filter by: CRITICAL, MAJOR, MINOR or WARNING. Must be upper-case.")
	cmd.Flags().Bool("resolved", false, "When set to true only resolved alarms will be updated (the one with status CLEARED), false means alarms with status ACTIVE or ACKNOWLEDGED.")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of alarm occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of alarm occurrence.")
	cmd.Flags().String("newStatus", "", "New status to be applied to all of the matching alarms")
	cmd.Flags().String("createdFrom", "", "Start date or date and time of the alarm creation.")
	cmd.Flags().String("createdTo", "", "End date or date and time of the alarm creation.")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "ACTIVE", "ACKNOWLEDGED", "CLEARED"),
		completion.WithValidateSet("newStatus", "ACTIVE", "ACKNOWLEDGED", "CLEARED"),
		completion.WithValidateSet("severity", "CRITICAL", "MAJOR", "MINOR", "WARNING"),
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("device", "source.id", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Update-AlarmCollection"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("device"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("newStatus", "status"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("status"),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := alarms.BulkUpdateOptions{
			DateFrom:    in.TimeValue("dateFrom"),
			DateTo:      in.TimeValue("dateTo"),
			CreatedFrom: in.TimeValue("createdFrom"),
			CreatedTo:   in.TimeValue("createdTo"),
			Resolved:    in.Bool("resolved"),
		}
		if s := in.String("severity"); s != "" {
			opt.Severity = []model.AlarmSeverity{model.AlarmSeverity(s)}
		}
		if st := in.String("status"); st != "" {
			opt.Status = []model.AlarmStatus{model.AlarmStatus(st)}
		}
		if device := in.String("device"); device != "" {
			opt.Source = managedobjects.DeviceRef(c8ystream.NameOrID(device))
		}
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Alarm] {
				return client.Alarms.UpdateList(ctx, opt, body)
			})
		}, nil
	})
}
