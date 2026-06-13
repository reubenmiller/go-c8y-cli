// Prototype of a v2-based create command: the body builder (data flag,
// templates, piped "name" values) is evaluated once per input item, and each
// resolved body becomes a typed Devices.Create call executed on the shared
// worker pool. Results stream through the same output pipeline as list
// commands, so --filter/--select/--outputTemplate/--outputFile work on
// created objects too.
package create

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// Create2Cmd command
type Create2Cmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreate2Cmd creates a command to Create device
func NewCreate2Cmd(f *cmdutil.Factory) *Create2Cmd {
	ccmd := &Create2Cmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create2",
		Short: "Create device",
		Long: `Create a device (managed object) with the special c8y_IsDevice fragment.
`,
		Example: heredoc.Doc(`
$ c8y devices create2 --name myDevice
Create device

$ c8y devices create2 --name myDevice --data "custom_value1=1234"
Create device with custom properties

$ echo -e "device01\ndevice02" | c8y devices create2 --workers 2
Create two devices (concurrently) using the piped names
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Device name (accepts pipeline)")
	cmd.Flags().String("type", "", "Device type")

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("name", "name", false, "name"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *Create2Cmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	// body: bound pipeline values (name) advance once per input item
	err = r.Bind(c8ystream.Body(
		flags.WithOverrideValue("name", "name"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("type", "type"),
		flags.WithRequiredTemplateString(`
{c8y_IsDevice: {}}`),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("name"),
	))
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(ctx context.Context, args c8ystream.Args) output.Seq {
		return c8ystream.FromResult(client.Devices.Create(ctx, args.Body))
	})
}
