// Prototype of a v2-based update command: per input item the id template is
// resolved (including the device name -> id lookup) and the body builder is
// evaluated with that item, then both feed a typed Devices.Update call on the
// shared worker pool.
package update

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// Update2Cmd command
type Update2Cmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUpdate2Cmd creates a command to Update device
func NewUpdate2Cmd(f *cmdutil.Factory) *Update2Cmd {
	ccmd := &Update2Cmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "update2",
		Short: "Update device",
		Long:  `Update properties of an existing device`,
		Example: heredoc.Doc(`
$ c8y devices update2 --id 12345 --newName "MyDevice"
Update device by id

$ c8y devices update2 --id 12345 --template "{c8y_SupportedOperations:['c8y_Restart', 'c8y_Command']}"
Update device using a template

$ c8y devices list | c8y devices update2 --data "myFragment={}" --workers 5
Update all piped devices concurrently
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Device ID (required) (accepts pipeline)")
	cmd.Flags().String("newName", "", "Device name")

	completion.WithOptions(
		cmd,
		completion.WithDevice("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("id", "deviceId", "source.id", "managedObject.id", "id"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *Update2Cmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	// id: resolved per input item (including device name -> id lookup) and
	// handed to the typed service call instead of a URL path template.
	// body: evaluated once per input item with the item bound as input.
	err = r.Bind(
		c8ystream.Param("id",
			c8yfetcher.WithDeviceByNameFirstMatch(n.factory, args, "id", "id"),
		),
		c8ystream.Body(
			flags.WithDataFlagValue(),
			flags.WithStringValue("newName", "name"),
			cmdutil.WithTemplateValue(n.factory),
			flags.WithTemplateVariablesValue(),
		),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(ctx context.Context, args c8ystream.Args) output.Seq {
		return c8ystream.FromResult(client.Devices.Update(ctx, args.Value("id"), args.Body))
	})
}
