// v2-based bulk operations create: the CLI builds the full body (operation
// prototype, schedule), then the device-group reference the body carries
// (groupId) is resolved (name -> id) by go-c8y as part of the
// create call. A bulk operation targets a single group, so the group flag is a
// scalar string (a body getter cannot bind a string slice).

package bulkoperations

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/spf13/cobra"
)

// CreateCmd command
type CreateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateCmd creates a command to Create bulk operation
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create bulk operation",
		Long:  `Create a new bulk operation`,
		Example: heredoc.Doc(`
$ c8y bulkoperations create --group 1234 --startDate "60s" --creationRampSec 15 --operation "c8y_Restart={}"
Create bulk operation for a group

$ c8y devicegroups get --id 12345 | c8y bulkoperations create --startDate "10s" --creationRampSec 15 --operation "c8y_Restart={}"
Create bulk operation for a group (using pipeline)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{}, "Identifies the target group on which this operation should be performed. (accepts pipeline) Multiple groups fan out one bulk operation per group.")
	cmd.Flags().String("startDate", "", "Time when operations should be created. Defaults to 300s")
	cmd.Flags().Float32("creationRampSec", 0, "Delay between every operation creation.")
	cmd.Flags().String("operation", "", "Operation prototype to send to each device in the group")

	completion.WithOptions(
		cmd,
		completion.WithDeviceGroup("group", ccmd.factory.ClientV2),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("group", "groupId", false, "id"),
		flags.WithPipelineAliases("group", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("New-BulkOperation"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.bulkoperation+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.Input(); err != nil {
		return err
	}

	err = r.Body(
		flags.WithOverrideValue("group", "groupId"),
		flags.WithDataFlagValue(),
		flags.WithSliceRefValue("group", "groupId"),
		flags.WithRelativeTimestamp("startDate", "startDate"),
		flags.WithFloatValue("creationRampSec", "creationRamp"),
		flags.WithDataValue("operation", "operationPrototype"),
		flags.WithDefaultTemplateString(`
{startDate: _.Now('300s'), creationRamp: 1.000}`),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("groupId", "startDate", "creationRamp", "operationPrototype"),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		// Restore v1's multi-group fan-out: --group 1,2 zips one group per piped
		// item (cycling), overriding the groupId seeded by WithOverrideValue.
		if body, err = in.ApplyOverrideSlice(body, "group", "groupId"); err != nil {
			return nil, err
		}
		return func(ctx context.Context) any {
			return client.BulkOperations.Create(ctx, body)
		}, nil
	})
}
