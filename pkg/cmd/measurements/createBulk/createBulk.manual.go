package createBulk

import (
	"context"
	"encoding/json"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/spf13/cobra"
)

// CreateBulkCmd command
type CreateBulkCmd struct {
	*subcommand.SubCommand

	batchSize int

	factory *cmdutil.Factory
}

// CreateBulkCmd creates a command to Create multiple measurements via the bulk measurement api
func NewCreateBulkCmd(f *cmdutil.Factory) *CreateBulkCmd {
	ccmd := &CreateBulkCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "createBulk",
		Short: "Create bulk measurements",
		Long: heredoc.Doc(`
			Create new measurements using bulk api

			It collects inputs and groups them into batches and send them to the server.

			Notes:
			* The template is applied to each piped measurement before being grouped
		`),
		Example: heredoc.Doc(`
			$ c8y measurements list -p 1000 --device 11111  | c8y measurements createBulk --device 22222 --batchSize 100
			Copy measurements from one device to another, but create measurement in batches of 100

			$ c8y measurements list -p 20 --valueFragmentType c8y_Temperature --valueFragmentSeries T \
				| c8y measurements createBulk \
					--type testme \
					--template "{c8y_Temperature+:{T+:{value: input.value.c8y_Temperature.T.value + 100}}}"
			Copy measurements from one device to another modifying the measurements slightly but adding 100 to the value
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "The ManagedObject which is the source of this measurement. (accepts pipeline)")
	cmd.Flags().String("time", "", "Time of the measurement. Defaults to current timestamp.")
	cmd.Flags().String("type", "", "The most specific type of this entire measurement.")
	cmd.Flags().IntVar(&ccmd.batchSize, "batchSize", 10, "Batch size. Number of measurements per request to send")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", ccmd.factory.ClientV2),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("device", "source.id", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateBulkCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	if err := r.Input(); err != nil {
		return err
	}
	err = r.Body(
		flags.WithOverrideValue("device", "source.id"),
		flags.WithDataFlagValue(),
		flags.WithSliceRefValue("device", "source.id"),
		flags.WithRelativeTimestamp("time", "time", ""),
		flags.WithStringValue("type", "type"),
		flags.WithDefaultTemplateString(`{time: _.Now('0s')} + (if std.isObject(input.value) then input.value else {source:{id:input.value}}) + {id:: '', 'self':: ''}`),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("type", "time", "source.id"),
	)
	if err != nil {
		return err
	}
	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.RunBatch(n.batchSize,
		func(in *c8ystream.Resolver) (json.RawMessage, error) {
			body, err := in.Body()
			if err != nil {
				return nil, err
			}
			// One device for every measurement (the first --device value). The
			// SDK resolves a device name
			return in.ApplyOverrideFirst(body, "device", "source.id")
		},
		func(items []json.RawMessage) c8ystream.Call {
			return func(ctx context.Context) any {
				return client.Measurements.CreateList(ctx, map[string]any{
					"measurements": items,
				})
			}
		},
	)
}
